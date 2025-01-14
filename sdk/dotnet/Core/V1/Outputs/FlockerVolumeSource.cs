// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.Kubernetes.Types.Outputs.Core.V1
{

    /// <summary>
    /// Represents a Flocker volume mounted by the Flocker agent. One and only one of datasetName and datasetUUID should be set. Flocker volumes do not support ownership management or SELinux relabeling.
    /// </summary>
    [OutputType]
    public sealed class FlockerVolumeSource
    {
        /// <summary>
        /// Name of the dataset stored as metadata -&gt; name on the dataset for Flocker should be considered as deprecated
        /// </summary>
        public readonly string DatasetName;
        /// <summary>
        /// UUID of the dataset. This is unique identifier of a Flocker dataset
        /// </summary>
        public readonly string DatasetUUID;

        [OutputConstructor]
        private FlockerVolumeSource(
            string datasetName,

            string datasetUUID)
        {
            DatasetName = datasetName;
            DatasetUUID = datasetUUID;
        }
    }
}
