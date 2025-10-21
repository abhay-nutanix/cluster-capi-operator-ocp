/*
Copyright 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package mapi2capi

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"

	"github.com/openshift/cluster-capi-operator/pkg/conversion/test/matchers"

	mapiv1 "github.com/openshift/api/machine/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	machinebuilder "github.com/openshift/cluster-api-actuator-pkg/testutils/resourcebuilder/machine/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// nutanixProviderSpecBuilder helps build Nutanix provider specs for testing.
type nutanixProviderSpecBuilder struct {
	providerSpec *mapiv1.NutanixMachineProviderConfig
}

// NutanixProviderSpec creates a new Nutanix machine config builder with a clean base configuration.
func NutanixProviderSpec() nutanixProviderSpecBuilder {
	return nutanixProviderSpecBuilder{
		providerSpec: &mapiv1.NutanixMachineProviderConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "machine.openshift.io/v1",
				Kind:       "NutanixMachineProviderConfig",
			},
			VCPUSockets:    2,
			VCPUsPerSocket: 1,
			MemorySize:     resource.MustParse("4Gi"),
			SystemDiskSize: resource.MustParse("120Gi"),
			Image: mapiv1.NutanixResourceIdentifier{
				Type: mapiv1.NutanixIdentifierName,
				Name: ptr.To("test-image"),
			},
			Subnets: []mapiv1.NutanixResourceIdentifier{
				{
					Type: mapiv1.NutanixIdentifierUUID,
					UUID: ptr.To("subnet-uuid"),
				},
			},
			Cluster: mapiv1.NutanixResourceIdentifier{
				Type: mapiv1.NutanixIdentifierUUID,
				UUID: ptr.To("cluster-uuid"),
			},
			Project: mapiv1.NutanixResourceIdentifier{
				Type: mapiv1.NutanixIdentifierUUID,
				UUID: ptr.To("project-uuid"),
			},
			BootType: mapiv1.NutanixLegacyBoot,
		},
	}
}

// NutanixProviderSpecClean creates a clean base spec without any problematic fields.
func NutanixProviderSpecClean() nutanixProviderSpecBuilder {
	return NutanixProviderSpec()
}

// Build returns the built provider spec.
func (n nutanixProviderSpecBuilder) Build() *mapiv1.NutanixMachineProviderConfig {
	return n.providerSpec.DeepCopy()
}

// BuildRawExtension returns the provider spec as a RawExtension.
func (n nutanixProviderSpecBuilder) BuildRawExtension() *runtime.RawExtension {
	objBytes, err := yaml.Marshal(n.providerSpec)
	if err != nil {
		panic(err)
	}

	return &runtime.RawExtension{
		Raw: objBytes,
	}
}

// WithImage sets the image for the machine.
func (n nutanixProviderSpecBuilder) WithImage(image mapiv1.NutanixResourceIdentifier) nutanixProviderSpecBuilder {
	n.providerSpec.Image = image
	return n
}

// WithSubnets sets the subnets for the machine.
func (n nutanixProviderSpecBuilder) WithSubnets(subnets []mapiv1.NutanixResourceIdentifier) nutanixProviderSpecBuilder {
	n.providerSpec.Subnets = subnets
	return n
}

// WithCluster sets the cluster for the machine.
func (n nutanixProviderSpecBuilder) WithCluster(cluster mapiv1.NutanixResourceIdentifier) nutanixProviderSpecBuilder {
	n.providerSpec.Cluster = cluster
	return n
}

// WithDataDisks sets the data disks for the machine.
func (n nutanixProviderSpecBuilder) WithDataDisks(dataDisks []mapiv1.NutanixVMDisk) nutanixProviderSpecBuilder {
	n.providerSpec.DataDisks = dataDisks
	return n
}

// WithGPUs sets the GPUs for the machine.
func (n nutanixProviderSpecBuilder) WithGPUs(gpus []mapiv1.NutanixGPU) nutanixProviderSpecBuilder {
	n.providerSpec.GPUs = gpus
	return n
}

// WithBootType sets the boot type for the machine.
func (n nutanixProviderSpecBuilder) WithBootType(bootType mapiv1.NutanixBootType) nutanixProviderSpecBuilder {
	n.providerSpec.BootType = bootType
	return n
}

// WithProject sets the project for the machine.
func (n nutanixProviderSpecBuilder) WithProject(project mapiv1.NutanixResourceIdentifier) nutanixProviderSpecBuilder {
	n.providerSpec.Project = project
	return n
}

// Note: WithCategories and WithFailureDomain methods are omitted as these types are not available in the test environment

var _ = Describe("mapi2capi Nutanix conversion", func() {
	var (
		nutanixBaseProviderSpec   = NutanixProviderSpecClean()
		nutanixMAPIMachineBase    = machinebuilder.Machine().WithProviderSpecBuilder(nutanixBaseProviderSpec)
		nutanixMAPIMachineSetBase = machinebuilder.MachineSet().WithProviderSpecBuilder(nutanixBaseProviderSpec)

		infra = &configv1.Infrastructure{
			Spec:   configv1.InfrastructureSpec{},
			Status: configv1.InfrastructureStatus{InfrastructureName: "sample-cluster-name"},
		}
	)

	type nutanixMAPI2CAPIConversionInput struct {
		machineBuilder   machinebuilder.MachineBuilder
		infra            *configv1.Infrastructure
		expectedErrors   []string
		expectedWarnings []string
	}

	type nutanixMAPI2CAPIMachinesetConversionInput struct {
		machineSetBuilder machinebuilder.MachineSetBuilder
		infra             *configv1.Infrastructure
		expectedErrors    []string
		expectedWarnings  []string
	}

	_ = DescribeTable("mapi2capi Nutanix convert MAPI Machine",
		func(in nutanixMAPI2CAPIConversionInput) {
			_, _, warns, err := FromNutanixMachineAndInfra(in.machineBuilder.Build(), in.infra).ToMachineAndInfrastructureMachine()
			Expect(err).To(matchers.ConsistOfMatchErrorSubstrings(in.expectedErrors), "should match expected errors while converting a Nutanix MAPI Machine to CAPI")
			Expect(warns).To(matchers.ConsistOfSubstrings(in.expectedWarnings), "should match expected warnings while converting a Nutanix MAPI Machine to CAPI")
		},

		// Base Case.
		Entry("With a Base configuration", nutanixMAPI2CAPIConversionInput{
			machineBuilder:   nutanixMAPIMachineBase,
			infra:            infra,
			expectedErrors:   []string{},
			expectedWarnings: []string{},
		}),

		// Error Cases - Resource Identifier Validation
		Entry("fails with Name type identifier missing name", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithImage(mapiv1.NutanixResourceIdentifier{
					Type: mapiv1.NutanixIdentifierName,
					Name: nil, // Missing name
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"name: Required value: Name must be set for Name type identifier",
			},
			expectedWarnings: []string{},
		}),

		Entry("fails with UUID type identifier missing UUID", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithCluster(mapiv1.NutanixResourceIdentifier{
					Type: mapiv1.NutanixIdentifierUUID,
					UUID: nil, // Missing UUID
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"uuid: Required value: UUID must be set for UUID type identifier",
			},
			expectedWarnings: []string{},
		}),

		Entry("fails with invalid identifier type", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithSubnets([]mapiv1.NutanixResourceIdentifier{
					{
						Type: "invalid-type", // Invalid type
						Name: ptr.To("subnet-name"),
					},
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"type: Invalid value: \"invalid-type\": invalid identifier type",
			},
			expectedWarnings: []string{},
		}),

		// Error Cases - Boot Type Validation
		Entry("fails with invalid boot type", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithBootType("invalid-boot-type"),
			),
			infra: infra,
			expectedErrors: []string{
				"bootType: Invalid value: \"invalid-boot-type\": invalid boot type",
			},
			expectedWarnings: []string{},
		}),

		// Error Cases - GPU Validation
		Entry("fails with GPU DeviceID type missing deviceID", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithGPUs([]mapiv1.NutanixGPU{
					{
						Type:     mapiv1.NutanixGPUIdentifierDeviceID,
						DeviceID: nil, // Missing DeviceID
					},
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"gpus[0]: Required value: DeviceID must be set for DeviceID type GPU",
			},
			expectedWarnings: []string{},
		}),

		Entry("fails with GPU Name type missing name", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithGPUs([]mapiv1.NutanixGPU{
					{
						Type: mapiv1.NutanixGPUIdentifierName,
						Name: nil, // Missing Name
					},
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"gpus[0]: Required value: Name must be set for Name type GPU",
			},
			expectedWarnings: []string{},
		}),

		Entry("fails with invalid GPU identifier type", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithGPUs([]mapiv1.NutanixGPU{
					{
						Type: "invalid-gpu-type", // Invalid type
						Name: ptr.To("gpu-name"),
					},
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"gpus[0]: Invalid value: \"invalid-gpu-type\": invalid GPU identifier type",
			},
			expectedWarnings: []string{},
		}),

		// Error Cases - VMDisk Validation
		Entry("fails with invalid disk device type", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithDataDisks([]mapiv1.NutanixVMDisk{
					{
						DiskSize: resource.MustParse("10Gi"),
						DeviceProperties: &mapiv1.NutanixVMDiskDeviceProperties{
							DeviceType: "InvalidDeviceType", // Invalid device type
						},
					},
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"DeviceType: Invalid value: \"InvalidDeviceType\": DeviceType should be CDRom or Disk",
				"AdapterType: Invalid value: \"\": AdapterType can be SCSI, IDE, PCI, SATA or SPAPR",
			},
			expectedWarnings: []string{},
		}),

		Entry("fails with invalid disk adapter type", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithDataDisks([]mapiv1.NutanixVMDisk{
					{
						DiskSize: resource.MustParse("10Gi"),
						DeviceProperties: &mapiv1.NutanixVMDiskDeviceProperties{
							AdapterType: "InvalidAdapterType", // Invalid adapter type
						},
					},
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"DeviceType: Invalid value: \"\": DeviceType should be CDRom or Disk",
				"AdapterType: Invalid value: \"InvalidAdapterType\": AdapterType can be SCSI, IDE, PCI, SATA or SPAPR",
			},
			expectedWarnings: []string{},
		}),

		Entry("fails with invalid disk mode", nutanixMAPI2CAPIConversionInput{
			machineBuilder: machinebuilder.Machine().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithDataDisks([]mapiv1.NutanixVMDisk{
					{
						DiskSize: resource.MustParse("10Gi"),
						StorageConfig: &mapiv1.NutanixVMStorageConfig{
							DiskMode: "InvalidDiskMode", // Invalid disk mode
						},
					},
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"DiskMode: Invalid value: \"InvalidDiskMode\": DiskMode can be Standard and Flash",
			},
			expectedWarnings: []string{},
		}),

		// Error Cases - Infrastructure validation
		Entry("fails with nil infrastructure", nutanixMAPI2CAPIConversionInput{
			machineBuilder:   nutanixMAPIMachineBase,
			infra:            nil,
			expectedErrors:   []string{"infrastructure cannot be nil and infrastructure.Status.InfrastructureName cannot be empty"},
			expectedWarnings: []string{},
		}),

		Entry("fails with empty infrastructure name", nutanixMAPI2CAPIConversionInput{
			machineBuilder: nutanixMAPIMachineBase,
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{InfrastructureName: ""},
			},
			expectedErrors:   []string{"infrastructure cannot be nil and infrastructure.Status.InfrastructureName cannot be empty"},
			expectedWarnings: []string{},
		}),
	)

	_ = DescribeTable("mapi2capi Nutanix convert MAPI MachineSet",
		func(in nutanixMAPI2CAPIMachinesetConversionInput) {
			_, _, warns, err := FromNutanixMachineSetAndInfra(in.machineSetBuilder.Build(), in.infra).ToMachineSetAndMachineTemplate()
			Expect(err).To(matchers.ConsistOfMatchErrorSubstrings(in.expectedErrors), "should match expected errors while converting a Nutanix MAPI MachineSet to CAPI")
			Expect(warns).To(matchers.ConsistOfSubstrings(in.expectedWarnings), "should match expected warnings while converting a Nutanix MAPI MachineSet to CAPI")
		},

		Entry("With a Base configuration", nutanixMAPI2CAPIMachinesetConversionInput{
			machineSetBuilder: nutanixMAPIMachineSetBase,
			infra:             infra,
			expectedErrors:    []string{},
			expectedWarnings:  []string{},
		}),

		Entry("fails with invalid resource identifier in MachineSet", nutanixMAPI2CAPIMachinesetConversionInput{
			machineSetBuilder: machinebuilder.MachineSet().WithProviderSpecBuilder(
				NutanixProviderSpecClean().WithImage(mapiv1.NutanixResourceIdentifier{
					Type: "invalid-type",
					Name: ptr.To("image-name"),
				}),
			),
			infra: infra,
			expectedErrors: []string{
				"type: Invalid value: \"invalid-type\": invalid identifier type",
			},
			expectedWarnings: []string{},
		}),

		Entry("fails with nil infrastructure in MachineSet", nutanixMAPI2CAPIMachinesetConversionInput{
			machineSetBuilder: nutanixMAPIMachineSetBase,
			infra:             nil,
			expectedErrors: []string{
				"infrastructure cannot be nil and infrastructure.Status.InfrastructureName cannot be empty",
				"infrastructure cannot be nil and infrastructure.Status.InfrastructureName cannot be empty",
			},
			expectedWarnings: []string{},
		}),
	)
})
