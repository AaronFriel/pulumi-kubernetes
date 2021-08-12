package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/mapstructure"
	"github.com/pulumi/pulumi-kubernetes/provider/v3/pkg/metadata"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	logger "github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	helmchart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/strvals"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// errReleaseNotFound is the error when a Helm release is not found
var errReleaseNotFound = errors.New("release not found")

type Release struct {
	ResourceType string       `json:"resourceType,omitempty"`
	ReleaseSpec  *ReleaseSpec `json:"releaseSpec,omitempty"`
	// Status of the deployed release.
	Status *ReleaseStatus `json:"status,omitempty"`
}

type ReleaseSpec struct {
	// If set, installation process purges chart on fail. The wait flag will be set automatically if atomic is used
	Atomic bool `json:"atomic,omitempty"`
	// Chart name to be installed. A path may be used.
	Chart string `json:"chart,omitempty"`
	// Allow deletion of new resources created in this upgrade when upgrade fails
	CleanupOnFail bool `json:"cleanupOnFail,omitempty"`
	// Create the namespace if it does not exist
	CreateNamespace bool `json:"createNamespace,omitempty"`
	// Run helm dependency update before installing the chart
	DependencyUpdate bool `json:"dependencyUpdate,omitempty"`
	// Add a custom description
	Description string `json:"description,omitempty"`
	// Use chart development versions, too. Equivalent to version '>0.0.0-0'. If `version` is set, this is ignored
	Devel bool `json:"devel,omitempty"`
	// Prevent CRD hooks from, running, but run other hooks.  See helm install --no-crd-hook
	DisableCRDHooks bool `json:"disableCRDHooks,omitempty"`
	// If set, the installation process will not validate rendered templates against the Kubernetes OpenAPI Schema
	DisableOpenapiValidation bool `json:"disableOpenapiValidation,omitempty"`
	// Prevent hooks from running.
	DisableWebhooks bool `json:"disableWebhooks,omitempty"`
	// Force resource update through delete/recreate if needed.
	ForceUpdate bool `json:"forceUpdate,omitempty"`
	// Location of public keys used for verification. Used only if `verify` is true
	Keyring string `json:"keyring,omitempty"`
	// Run helm lint when planning
	Lint bool `json:"lint,omitempty"`
	// Limit the maximum number of revisions saved per release. Use 0 for no limit
	MaxHistory *int `json:"maxHistory,omitempty"`
	// Release name.
	Name string `json:"name,omitempty"`
	// Namespace to install the release into.
	Namespace string `json:"namespace,omitempty"`
	// Postrender command to run.
	Postrender string `json:"postrender,omitempty"`
	// Perform pods restart during upgrade/rollback
	RecreatePods bool `json:"recreatePods,omitempty"`
	// If set, render subchart notes along with the parent
	RenderSubchartNotes bool `json:"renderSubchartNotes,omitempty"`
	// Re-use the given name, even if that name is already used. This is unsafe in production
	Replace bool `json:"replace,omitempty"`
	// Specification defining the Helm chart repository to use.
	RepositorySpec RepositorySpec `json:"repositorySpec,omitempty"`
	// When upgrading, reset the values to the ones built into the chart
	ResetValues bool `json:"resetValues,omitempty"`
	// When upgrading, reuse the last release's values and merge in any overrides. If 'reset_values' is specified, this is ignored
	ReuseValues bool `json:"reuseValues,omitempty"`
	// Custom values to be merged with the values.
	Set []*SetValue `json:"set,omitempty"`
	// If set, no CRDs will be installed. By default, CRDs are installed if not already present
	SkipCrds bool `json:"skipCrds,omitempty"`
	// Time in seconds to wait for any individual kubernetes operation.
	Timeout int `json:"timeout,omitempty"`
	// List of values in raw yaml format to pass to helm.
	Values []string `json:"values,omitempty"`
	// Verify the package before installing it.
	Verify bool `json:"verify,omitempty"`
	// Specify the exact chart version to install. If this is not specified, the latest version is installed.
	Version string `json:"version,omitempty"`
	// Will wait until all resources are in a ready state before marking the release as successful.
	Wait bool `json:"wait,omitempty"`
	// If wait is enabled, will wait until all Jobs have been completed before marking the release as successful.
	WaitForJobs bool `json:"waitForJobs,omitempty"`
}

// Specification defining the Helm chart repository to use.
type RepositorySpec struct {
	// Repository where to locate the requested chart. If is a URL the chart is installed without installing the repository.
	Repository string `json:"repository,omitempty"`
	// The Repositories CA File
	RepositoryCAFile string `json:"repositoryCAFile,omitempty"`
	// The repositories cert file
	RepositoryCertFile string `json:"repositoryCertFile,omitempty"`
	// The repositories cert key file
	RepositoryKeyFile string `json:"repositoryKeyFile,omitempty"`
	// Password for HTTP basic authentication
	RepositoryPassword string `json:"repositoryPassword,omitempty"`
	// Username for HTTP basic authentication
	RepositoryUsername string `json:"repositoryUsername,omitempty"`
}

type SetValue struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type ReleaseStatus struct {
	// The version number of the application being deployed.
	AppVersion string `json:"appVersion,omitempty"`
	// The name of the chart.
	Chart string `json:"chart,omitempty"`
	// Name is the name of the release.
	Name string `json:"name,omitempty"`
	// Namespace is the kubernetes namespace of the release.
	Namespace string `json:"namespace,omitempty"`
	// Version is an int32 which represents the version of the release.
	Revision *int `json:"revision,omitempty"`
	// Status of the release.
	Status string `json:"status,omitempty"`
	// Set of extra values, added to the chart. The sensitive data is cloaked. JSON encoded.
	Values string `json:"values,omitempty"`
	// A SemVer 2 conformant version string of the chart.
	Version string `json:"version,omitempty"`
	// The rendered manifest as JSON.
	Manifest string `json:"manifest,omitempty"`
}

type helmReleaseProvider struct {
	helmDriver       string
	kubeConfig       *KubeConfig
	defaultNamespace string
	enableSecrets    bool
	name             string
	settings         *cli.EnvSettings
}

func newHelmReleaseProvider(
	config *rest.Config,
	clientConfig clientcmd.ClientConfig,
	helmDriver,
	namespace string,
	enableSecrets bool,
	pluginsDirectory,
	registryConfigPath,
	repositoryConfigPath,
	repositoryCache string,
) (customResourceProvider, error) {
	kc := newKubeConfig(config, clientConfig)
	settings := cli.New()
	settings.PluginsDirectory = pluginsDirectory
	settings.RegistryConfig = registryConfigPath
	settings.RepositoryConfig = repositoryConfigPath
	settings.RepositoryCache = repositoryCache

	return &helmReleaseProvider{
		kubeConfig:       kc,
		helmDriver:       helmDriver,
		defaultNamespace: namespace,
		enableSecrets:    enableSecrets,
		name:             "kubernetes:helmrelease",
		settings:         settings,
	}, nil
}

func debug(format string, a ...interface{}) {
	logger.V(9).Infof("[DEBUG] %s", fmt.Sprintf(format, a...))
}

func (r *helmReleaseProvider) getActionConfig(namespace string) (*action.Configuration, error) {
	conf := new(action.Configuration)
	if err := conf.Init(r.kubeConfig, namespace, r.helmDriver, debug); err != nil {
		return nil, err
	}
	return conf, nil
}

func (r *helmReleaseProvider) Invoke(ctx context.Context, request *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	panic("implement me")
}

func (r *helmReleaseProvider) StreamInvoke(request *pulumirpc.InvokeRequest, server pulumirpc.ResourceProvider_StreamInvokeServer) error {
	panic("implement me")
}

func decodeRelease(pm resource.PropertyMap) (*Release, error) {
	var release Release
	stripped := pm.MapRepl(nil, mapReplStripSecrets)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &release,
		TagName: "json",
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(stripped); err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *helmReleaseProvider) Check(ctx context.Context, req *pulumirpc.CheckRequest, olds, news resource.PropertyMap) (*pulumirpc.CheckResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("Provider[%s].Check(%s)", r.name, urn)

	new, err := decodeRelease(news)
	if err != nil {
		return nil, err
	}

	if len(olds.Mappable()) > 0 {
		old, err := decodeRelease(olds)
		if err != nil {
			return nil, err
		}
		adoptOldNameIfUnnamed(new, old)

		if new.ReleaseSpec.Namespace == "" {
			new.ReleaseSpec.Namespace = old.ReleaseSpec.Namespace
		}
	} else {
		assignNameIfAutonammable(new, news, "release")
	}

	if new.ReleaseSpec.Namespace == "" {
		new.ReleaseSpec.Namespace = r.defaultNamespace
	}

	autonamed := resource.NewPropertyMap(new)
	annotateSecrets(autonamed, news)
	autonamedInputs, err := plugin.MarshalProperties(autonamed, plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.autonamedInputs", label),
		KeepUnknowns: true,
		SkipNulls:    true,
		KeepSecrets:  r.enableSecrets,
	})
	if err != nil {
		return nil, err
	}

	// Return new, possibly-autonamed inputs.
	return &pulumirpc.CheckResponse{Inputs: autonamedInputs}, nil
}

func adoptOldNameIfUnnamed(new, old *Release) {
	contract.Assert(old.ReleaseSpec.Name != "")
	new.ReleaseSpec.Name = old.ReleaseSpec.Name
}

func assignNameIfAutonammable(release *Release, pm resource.PropertyMap, base tokens.QName) {
	if rs, ok := pm["resourceSpec"].V.(resource.PropertyMap); ok {
		if name, ok := rs["name"]; ok && name.IsComputed() {
			return
		}
		if name, ok := rs["name"]; !ok || name.StringValue() == "" {
			release.ReleaseSpec.Name = fmt.Sprintf("%s-%s", base, metadata.RandString(8))
		}
	}
}

func (r *helmReleaseProvider) Diff(ctx context.Context, request *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	panic("implement me")
}

func (r *helmReleaseProvider) Create(ctx context.Context, req *pulumirpc.CreateRequest, news resource.PropertyMap) (*pulumirpc.CreateResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("Provider[%s].Create(%s)", r.name, urn)

	newRelease, err := decodeRelease(news)
	if err != nil {
		return nil, err
	}
	// Freeze inputs to track the actual inputs for checkpointing.
	inputs, err := decodeRelease(news)
	contract.AssertNoError(err)

	conf, err := r.getActionConfig(newRelease.ReleaseSpec.Namespace)
	if err != nil {
		return nil, err
	}
	client := action.NewInstall(conf)
	cpo, chartName, err := chartPathOptions(newRelease.ReleaseSpec)

	c, path, err := getChart(chartName, r.settings, cpo)
	if err != nil {
		return nil, err
	}

	// check and update the chart's dependencies if needed
	updated, err := checkChartDependencies(c, path, newRelease.ReleaseSpec.Keyring, r.settings, newRelease.ReleaseSpec.DependencyUpdate)
	if err != nil {
		return nil, err
	} else if updated {
		// load the chart again if its dependencies have been updated
		c, err = loader.Load(path)
		if err != nil {
			return nil, err
		}
	}

	values, err := getValues(newRelease.ReleaseSpec)
	if err != nil {
		return nil, err
	}

	err = isChartInstallable(c)
	if err != nil {
		return nil, err
	}

	client.ChartPathOptions = *cpo
	client.ClientOnly = false
	client.DryRun = false
	client.DisableHooks = newRelease.ReleaseSpec.DisableWebhooks
	client.Wait = newRelease.ReleaseSpec.Wait
	client.WaitForJobs = newRelease.ReleaseSpec.WaitForJobs
	client.Devel = newRelease.ReleaseSpec.Devel
	client.DependencyUpdate = newRelease.ReleaseSpec.DependencyUpdate
	client.Timeout = time.Duration(newRelease.ReleaseSpec.Timeout) * time.Second
	client.Namespace = newRelease.ReleaseSpec.Namespace
	client.ReleaseName = newRelease.ReleaseSpec.Name
	client.GenerateName = false
	client.NameTemplate = ""
	client.OutputDir = ""
	client.Atomic = newRelease.ReleaseSpec.Atomic
	client.SkipCRDs = newRelease.ReleaseSpec.SkipCrds
	client.SubNotes = newRelease.ReleaseSpec.RenderSubchartNotes
	client.DisableOpenAPIValidation = newRelease.ReleaseSpec.DisableOpenapiValidation
	client.Replace = newRelease.ReleaseSpec.Replace
	client.Description = newRelease.ReleaseSpec.Description
	client.CreateNamespace = newRelease.ReleaseSpec.CreateNamespace

	if cmd := newRelease.ReleaseSpec.Postrender; cmd != "" {
		pr, err := postrender.NewExec(cmd)

		if err != nil {
			return nil, err
		}

		client.PostRenderer = pr
	}

	rel, err := client.Run(c, values)
	if err != nil && rel == nil {
		return nil, err
	}

	if err != nil && rel != nil {
		actionConfig, err := r.getActionConfig(newRelease.ReleaseSpec.Namespace)
		if err != nil {
			return nil, err
		}
		exists, existsErr := resourceReleaseExists(newRelease.ReleaseSpec, actionConfig)

		if existsErr != nil {
			return nil, err
		}

		if !exists {
			return nil, err
		}

		//debug("%s Release was created but returned an error", logID)

		if err := setReleaseAttributes(newRelease, news, rel); err != nil {
			return nil, err
		}

		//return diag.Diagnostics{
		//	{
		//		Severity: diag.Warning,
		//		Summary:  fmt.Sprintf("Helm release %q was created but has a failed status. Use the `helm` command to investigate the error, correct it, then run Terraform again.", client.ReleaseName),
		//	},
		//	{
		//		Severity: diag.Error,
		//		Summary:  err.Error(),
		//	},
		//}
		// TODO: k.host.LogStatus instead.
		return nil, err

	}

	err = setReleaseAttributes(newRelease, news, rel)
	if err != nil {
		return nil, err
	}

	obj := checkpointRelease(inputs, newRelease, news)
	inputsAndComputed, err := plugin.MarshalProperties(
		obj, plugin.MarshalOptions{
			Label:        fmt.Sprintf("%s.inputsAndComputed", label),
			KeepUnknowns: true,
			SkipNulls:    true,
			KeepSecrets:  r.enableSecrets,
		})
	if err != nil {
		return nil, err
	}

	id := ""
	if !req.GetPreview() {
		id = newRelease.ReleaseSpec.Name
	}
	return &pulumirpc.CreateResponse{Id: id, Properties: inputsAndComputed}, nil
}

func checkpointRelease(inputs, live *Release, fromInputs resource.PropertyMap) resource.PropertyMap {
	object := resource.NewPropertyMap(live)
	inputsPM := resource.NewPropertyMap(inputs)

	annotateSecrets(object, fromInputs)
	annotateSecrets(inputsPM, fromInputs)

	object["__inputs"] = resource.NewObjectProperty(inputsPM)

	return object
}

func setReleaseAttributes(release *Release, news resource.PropertyMap, r *release.Release) error {
	release.Status.Version = r.Chart.Metadata.Version
	release.Status.Namespace = r.Namespace
	release.Status.Name = r.Name
	release.Status.Status = r.Info.Status.String()

	cloakSetValues(r.Config, news)
	values, err := json.Marshal(r.Config)
	if err != nil {
		return err
	}

	jsonManifest, err := convertYAMLManifestToJSON(r.Manifest)
	if err != nil {
		return err
	}
	manifest := redactSensitiveValues(jsonManifest, news)
	release.Status.Manifest = manifest

	release.Status.Name = r.Name
	release.Status.Namespace = r.Namespace
	release.Status.Revision = &r.Version
	release.Status.Chart = r.Chart.Metadata.Name
	release.Status.Version = r.Chart.Metadata.Version
	release.Status.AppVersion = r.Chart.Metadata.AppVersion
	release.Status.Values = string(values)
	return nil
}

func resourceReleaseExists(releaseSpec *ReleaseSpec, actionConfig *action.Configuration) (bool, error) {
	logger.V(9).Infof("[resourceReleaseExists: %s]", releaseSpec.Name)
	name := releaseSpec.Name
	_, err := getRelease(actionConfig, name)

	logger.V(9).Infof("[resourceReleaseExists: %s] Done", releaseSpec.Name)

	if err == nil {
		return true, nil
	}

	if err == errReleaseNotFound {
		return false, nil
	}

	return false, err
}

func getRelease(cfg *action.Configuration, name string) (*release.Release, error) {
	get := action.NewGet(cfg)
	debug("%s getRelease post action created", name)

	res, err := get.Run(name)
	debug("%s getRelease post run", name)

	if err != nil {
		debug("getRelease for %s errored", name)
		debug("%v", err)
		if strings.Contains(err.Error(), "release: not found") {
			return nil, errReleaseNotFound
		}

		debug("could not get release %s", err)

		return nil, err
	}

	debug("%s getRelease done", name)

	return res, nil
}

func isChartInstallable(ch *helmchart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func getValues(spec *ReleaseSpec) (map[string]interface{}, error) {
	base := map[string]interface{}{}

	for _, value := range spec.Values {
		if value == "" {
			continue
		}

		currentMap := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(value), &currentMap); err != nil {
			return nil, fmt.Errorf("---> %v %s", err, value)
		}

		base = mergeMaps(base, currentMap)
	}

	for _, set := range spec.Set {
		if err := getValue(base, set); err != nil {
			return nil, err
		}
	}

	//for _, set := range spec.SetSensitive {
	//	if err := getValue(base, set); err != nil {
	//		return nil, err
	//	}
	//}

	return base, logValues(base, spec)
}

func getValue(base map[string]interface{}, set *SetValue) error {
	name := set.Name
	value := set.Value
	valueType := set.Type

	switch valueType {
	case "auto", "":
		if err := strvals.ParseInto(fmt.Sprintf("%s=%s", name, value), base); err != nil {
			return fmt.Errorf("failed parsing key %q with value %s, %s", name, value, err)
		}
	case "string":
		if err := strvals.ParseIntoString(fmt.Sprintf("%s=%s", name, value), base); err != nil {
			return fmt.Errorf("failed parsing key %q with value %s, %s", name, value, err)
		}
	default:
		return fmt.Errorf("unexpected type: %s", valueType)
	}

	return nil
}

func logValues(values map[string]interface{}, spec *ReleaseSpec) error {
	// copy array to avoid change values by the cloak function.
	//asJSON, _ := json.Marshal(values)
	//var c map[string]interface{}
	//err := json.Unmarshal(asJSON, &c)
	//if err != nil {
	//	return err
	//}
	//
	//cloakSetValues(c, spec)
	//
	//y, err := yaml.Marshal(c)
	//if err != nil {
	//	return err
	//}
	//
	//log.Printf(
	//	"---[ values.yaml ]-----------------------------------\n%s\n",
	//	string(y),
	//)

	return nil
}

func cloakSetValues(config map[string]interface{}, pm resource.PropertyMap) {
	//if rs, ok := pm["resourceSpec"].V.(resource.PropertyMap); ok {
	//	if set, ok := rs["set"]; ok && set.ContainsSecrets() {
	//		set.SecretValue().Element
	//	}
	//}
	//
	//for _, raw := range d.Get("set_sensitive").(*schema.Set).List() {
	//	set := raw.(map[string]interface{})
	//	cloakSetValue(config, set["name"].(string))
	//}
}

const sensitiveContentValue = "(sensitive value)"

func cloakSetValue(values map[string]interface{}, valuePath string) {
	pathKeys := strings.Split(valuePath, ".")
	sensitiveKey := pathKeys[len(pathKeys)-1]
	parentPathKeys := pathKeys[:len(pathKeys)-1]

	m := values
	for _, key := range parentPathKeys {
		v, ok := m[key].(map[string]interface{})
		if !ok {
			return
		}
		m = v
	}

	m[sensitiveKey] = sensitiveContentValue
}

// Merges source and destination map, preferring values from the source map
// Taken from github.com/helm/pkg/cli/values/options.go
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func getChart(name string, settings *cli.EnvSettings, cpo *action.ChartPathOptions) (*helmchart.Chart, string, error) {
	//Load function blows up if accessed concurrently
	path, err := cpo.LocateChart(name, settings)
	if err != nil {
		return nil, "", err
	}

	c, err := loader.Load(path)
	if err != nil {
		return nil, "", err
	}

	return c, path, nil
}

func checkChartDependencies(c *helmchart.Chart, path, keyring string, settings *cli.EnvSettings, dependencyUpdate bool) (bool, error) {
	p := getter.All(settings)

	if req := c.Metadata.Dependencies; req != nil {
		err := action.CheckDependencies(c, req)
		if err != nil {
			if dependencyUpdate {
				man := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        path,
					Keyring:          keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				log.Println("[DEBUG] Downloading chart dependencies...")
				return true, man.Update()
			}
			return false, err
		}
		return false, err
	}
	log.Println("[DEBUG] Chart dependencies are up to date.")
	return false, nil
}

func chartPathOptions(releaseSpec *ReleaseSpec) (*action.ChartPathOptions, string, error) {
	chartName := releaseSpec.Chart

	repository := releaseSpec.RepositorySpec.Repository
	repositoryURL, chartName, err := resolveChartName(repository, strings.TrimSpace(chartName))

	if err != nil {
		return nil, "", err
	}
	version := getVersion(releaseSpec)

	return &action.ChartPathOptions{
		CaFile:   releaseSpec.RepositorySpec.RepositoryCAFile,
		CertFile: releaseSpec.RepositorySpec.RepositoryCertFile,
		KeyFile:  releaseSpec.RepositorySpec.RepositoryKeyFile,
		//Keyring:  d.Get("keyring").(string),
		RepoURL:  repositoryURL,
		Verify:   releaseSpec.Verify,
		Version:  version,
		Username: releaseSpec.RepositorySpec.RepositoryUsername,
		Password: releaseSpec.RepositorySpec.RepositoryPassword, // TODO: This should already be resolved.
	}, chartName, nil
}

func getVersion(releaseSpec *ReleaseSpec) (version string) {
	version = releaseSpec.Version

	if version == "" && releaseSpec.Devel {
		debug("setting version to >0.0.0-0")
		version = ">0.0.0-0"
	} else {
		version = strings.TrimSpace(version)
	}

	return
}

func resolveChartName(repository, name string) (string, string, error) {
	_, err := url.ParseRequestURI(repository)
	if err == nil {
		return repository, name, nil
	}

	if strings.Index(name, "/") == -1 && repository != "" {
		name = fmt.Sprintf("%s/%s", repository, name)
	}

	return "", name, nil
}

func (r *helmReleaseProvider) Read(ctx context.Context, request *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	panic("implement me")
}

func (r *helmReleaseProvider) Update(ctx context.Context, request *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	panic("implement me")
}

func (r *helmReleaseProvider) Delete(ctx context.Context, request *pulumirpc.DeleteRequest) (*empty.Empty, error) {
	panic("implement me")
}

func isHelmRelease(urn resource.URN) bool {
	return urn.Type() == "kubernetes:helm.sh/v3:Release"
}

func asHelmRelease(pm resource.PropertyMap) (*Release, error) {
	obj := pm.MapRepl(nil, mapReplStripSecrets)
	var release Release
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &release,
		TagName: "json",
	})
	contract.AssertNoError(err)
	if err := decoder.Decode(obj); err != nil {
		return nil, err
	}

	return &release, nil
}