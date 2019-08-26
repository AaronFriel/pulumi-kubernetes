// *** WARNING: this file was generated by the Pulumi Kubernetes codegen tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import { core } from "../..";
import * as inputs from "../../types/input";
import * as outputs from "../../types/output";
import { getVersion } from "../../version";

    /**
     * configuration of a horizontal pod autoscaler.
     */
    export class HorizontalPodAutoscaler extends pulumi.CustomResource {
      /**
       * APIVersion defines the versioned schema of this representation of an object. Servers should
       * convert recognized schemas to the latest internal value, and may reject unrecognized
       * values. More info:
       * https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
       */
      public readonly apiVersion: pulumi.Output<"autoscaling/v1">;

      /**
       * Kind is a string value representing the REST resource this object represents. Servers may
       * infer this from the endpoint the client submits requests to. Cannot be updated. In
       * CamelCase. More info:
       * https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
       */
      public readonly kind: pulumi.Output<"HorizontalPodAutoscaler">;

      /**
       * Standard object metadata. More info:
       * https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
       */
      public readonly metadata: pulumi.Output<outputs.meta.v1.ObjectMeta>;

      /**
       * behaviour of autoscaler. More info:
       * https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
       */
      public readonly spec: pulumi.Output<outputs.autoscaling.v1.HorizontalPodAutoscalerSpec>;

      /**
       * current information about the autoscaler.
       */
      public readonly status: pulumi.Output<outputs.autoscaling.v1.HorizontalPodAutoscalerStatus>;

      /**
       * Get the state of an existing `HorizontalPodAutoscaler` resource, as identified by `id`.
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
      public static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): HorizontalPodAutoscaler {
          return new HorizontalPodAutoscaler(name, undefined, { ...opts, id: id });
      }

      /** @internal */
      private static readonly __pulumiType = "kubernetes:autoscaling/v1:HorizontalPodAutoscaler";

      /**
       * Returns true if the given object is an instance of HorizontalPodAutoscaler.  This is designed to work even
       * when multiple copies of the Pulumi SDK have been loaded into the same process.
       */
      public static isInstance(obj: any): obj is HorizontalPodAutoscaler {
          if (obj === undefined || obj === null) {
              return false;
          }

          return obj["__pulumiType"] === HorizontalPodAutoscaler.__pulumiType;
      }

      /**
       * Create a autoscaling.v1.HorizontalPodAutoscaler resource with the given unique name, arguments, and options.
       *
       * @param name The _unique_ name of the resource.
       * @param args The arguments to use to populate this resource's properties.
       * @param opts A bag of options that control this resource's behavior.
       */
      constructor(name: string, args?: inputs.autoscaling.v1.HorizontalPodAutoscaler, opts?: pulumi.CustomResourceOptions) {
          const props: pulumi.Inputs = {};

          props["apiVersion"] = "autoscaling/v1";
          props["kind"] = "HorizontalPodAutoscaler";
          props["metadata"] = args && args.metadata || undefined;
          props["spec"] = args && args.spec || undefined;

          props["status"] = undefined;

          if (!opts) {
              opts = {};
          }

          if (!opts.version) {
              opts.version = getVersion();
          }
          super(HorizontalPodAutoscaler.__pulumiType, name, props, opts);
      }
    }
