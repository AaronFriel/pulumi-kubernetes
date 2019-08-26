// *** WARNING: this file was generated by the Pulumi Kubernetes codegen tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import { core } from "../..";
import * as inputs from "../../types/input";
import * as outputs from "../../types/output";
import { getVersion } from "../../version";

    /**
     * CSINode holds information about all CSI drivers installed on a node. CSI drivers do not need
     * to create the CSINode object directly. As long as they use the node-driver-registrar sidecar
     * container, the kubelet will automatically populate the CSINode object for the CSI driver as
     * part of kubelet plugin registration. CSINode has the same name as a node. If the object is
     * missing, it means either there are no CSI Drivers available on the node, or the Kubelet
     * version is low enough that it doesn't create this object. CSINode has an OwnerReference that
     * points to the corresponding node object.
     */
    export class CSINode extends pulumi.CustomResource {
      /**
       * APIVersion defines the versioned schema of this representation of an object. Servers should
       * convert recognized schemas to the latest internal value, and may reject unrecognized
       * values. More info:
       * https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
       */
      public readonly apiVersion: pulumi.Output<"storage.k8s.io/v1beta1">;

      /**
       * Kind is a string value representing the REST resource this object represents. Servers may
       * infer this from the endpoint the client submits requests to. Cannot be updated. In
       * CamelCase. More info:
       * https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
       */
      public readonly kind: pulumi.Output<"CSINode">;

      /**
       * metadata.name must be the Kubernetes node name.
       */
      public readonly metadata: pulumi.Output<outputs.meta.v1.ObjectMeta>;

      /**
       * spec is the specification of CSINode
       */
      public readonly spec: pulumi.Output<outputs.storage.v1beta1.CSINodeSpec>;

      /**
       * Get the state of an existing `CSINode` resource, as identified by `id`.
       * Typically this ID  is of the form [namespace]/[name]; if [namespace] is omitted, then (per
       * Kubernetes convention) the ID becomes default/[name].
       *
       * Pulumi will keep track of this resource using `name` as the Pulumi ID.
       *
       * @param name _Unique_ name used to register this resource with Pulumi.
       * @param id An ID for the Kubernetes resource to retrieve. Takes the form
       *  [namespace]/[name] or [name].
       * @param opts Uniquely specifies a CustomResource to select.
       */
      public static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): CSINode {
          return new CSINode(name, undefined, { ...opts, id: id });
      }

      /** @internal */
      private static readonly __pulumiType = "kubernetes:storage.k8s.io/v1beta1:CSINode";

      /**
       * Returns true if the given object is an instance of CSINode.  This is designed to work even
       * when multiple copies of the Pulumi SDK have been loaded into the same process.
       */
      public static isInstance(obj: any): obj is CSINode {
          if (obj === undefined || obj === null) {
              return false;
          }

          return obj["__pulumiType"] === CSINode.__pulumiType;
      }

      /**
       * Create a storage.v1beta1.CSINode resource with the given unique name, arguments, and options.
       *
       * @param name The _unique_ name of the resource.
       * @param args The arguments to use to populate this resource's properties.
       * @param opts A bag of options that control this resource's behavior.
       */
      constructor(name: string, args?: inputs.storage.v1beta1.CSINode, opts?: pulumi.CustomResourceOptions) {
          const props: pulumi.Inputs = {};
          props["spec"] = args && args.spec || undefined;

          props["apiVersion"] = "storage.k8s.io/v1beta1";
          props["kind"] = "CSINode";
          props["metadata"] = args && args.metadata || undefined;

          props["status"] = undefined;

          if (!opts) {
              opts = {};
          }

          if (!opts.version) {
              opts.version = getVersion();
          }
          super(CSINode.__pulumiType, name, props, opts);
      }
    }
