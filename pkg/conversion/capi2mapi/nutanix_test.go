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
package capi2mapi

import (
	nutanixv1 "github.com/nutanix-cloud-native/cluster-api-provider-nutanix/api/v1beta1"
	credentialTypes "github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterv1resourcebuilder "github.com/openshift/cluster-api-actuator-pkg/testutils/resourcebuilder/cluster-api/core/v1beta1"
	"github.com/openshift/cluster-capi-operator/pkg/conversion/test/matchers"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

var _ = Describe("capi2mapi Nutanix conversion", func() {
	// Helper function to create a base NutanixMachine
	createBaseNutanixMachine := func() *nutanixv1.NutanixMachine {
		return &nutanixv1.NutanixMachine{
			TypeMeta: metav1.TypeMeta{
				APIVersion: nutanixv1.GroupVersion.String(),
				Kind:       "NutanixMachine",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-nutanix-machine",
				Namespace: "test-namespace",
			},
			Spec: nutanixv1.NutanixMachineSpec{
				VCPUsPerSocket: 2,
				VCPUSockets:    1,
				MemorySize:     resource.MustParse("4Gi"),
				SystemDiskSize: resource.MustParse("120Gi"),
				Image: &nutanixv1.NutanixResourceIdentifier{
					Type: nutanixv1.NutanixIdentifierName,
					Name: ptr.To("rhcos-image"),
				},
				Cluster: nutanixv1.NutanixResourceIdentifier{
					Type: nutanixv1.NutanixIdentifierName,
					Name: ptr.To("prism-cluster"),
				},
				Subnets: []nutanixv1.NutanixResourceIdentifier{
					{
						Type: nutanixv1.NutanixIdentifierName,
						Name: ptr.To("vm-network"),
					},
				},
				BootType: nutanixv1.NutanixBootTypeLegacy,
			},
		}
	}

	// Helper function to create a base NutanixCluster
	createBaseNutanixCluster := func() *nutanixv1.NutanixCluster {
		return &nutanixv1.NutanixCluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: nutanixv1.GroupVersion.String(),
				Kind:       "NutanixCluster",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-nutanix-cluster",
				Namespace: "test-namespace",
			},
			Spec: nutanixv1.NutanixClusterSpec{
				PrismCentral: &credentialTypes.NutanixPrismEndpoint{
					Address: "prism-central.example.com",
					Port:    9440,
					CredentialRef: &credentialTypes.NutanixCredentialReference{
						Kind: credentialTypes.SecretKind,
						Name: "nutanix-credentials",
					},
				},
			},
		}
	}

	// Helper function to create a base NutanixMachineTemplate
	createBaseNutanixMachineTemplate := func() *nutanixv1.NutanixMachineTemplate {
		return &nutanixv1.NutanixMachineTemplate{
			TypeMeta: metav1.TypeMeta{
				APIVersion: nutanixv1.GroupVersion.String(),
				Kind:       "NutanixMachineTemplate",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-nutanix-machine-template",
				Namespace: "test-namespace",
			},
			Spec: nutanixv1.NutanixMachineTemplateSpec{
				Template: nutanixv1.NutanixMachineTemplateResource{
					Spec: nutanixv1.NutanixMachineSpec{
						VCPUsPerSocket: 2,
						VCPUSockets:    1,
						MemorySize:     resource.MustParse("4Gi"),
						SystemDiskSize: resource.MustParse("120Gi"),
						Image: &nutanixv1.NutanixResourceIdentifier{
							Type: nutanixv1.NutanixIdentifierName,
							Name: ptr.To("rhcos-image"),
						},
						Cluster: nutanixv1.NutanixResourceIdentifier{
							Type: nutanixv1.NutanixIdentifierName,
							Name: ptr.To("prism-cluster"),
						},
						Subnets: []nutanixv1.NutanixResourceIdentifier{
							{
								Type: nutanixv1.NutanixIdentifierName,
								Name: ptr.To("vm-network"),
							},
						},
						BootType: nutanixv1.NutanixBootTypeLegacy,
					},
				},
			},
		}
	}

	type nutanixCAPI2MAPIMachineConversionInput struct {
		machine          *clusterv1.Machine
		nutanixMachine   *nutanixv1.NutanixMachine
		nutanixCluster   *nutanixv1.NutanixCluster
		expectedErrors   []string
		expectedWarnings []string
	}

	type nutanixCAPI2MAPIMachinesetConversionInput struct {
		machineSet             *clusterv1.MachineSet
		nutanixMachineTemplate *nutanixv1.NutanixMachineTemplate
		nutanixCluster         *nutanixv1.NutanixCluster
		expectedErrors         []string
		expectedWarnings       []string
	}

	_ = DescribeTable("capi2mapi Nutanix convert CAPI Machine/InfraMachine/InfraCluster to a MAPI Machine",
		func(in nutanixCAPI2MAPIMachineConversionInput) {
			_, warns, err := FromMachineAndNutanixMachineAndNutanixCluster(
				in.machine,
				in.nutanixMachine,
				in.nutanixCluster,
			).ToMachine()
			Expect(err).To(matchers.ConsistOfMatchErrorSubstrings(in.expectedErrors),
				"should match expected errors while converting Nutanix CAPI resources to MAPI Machine")
			Expect(warns).To(matchers.ConsistOfSubstrings(in.expectedWarnings),
				"should match expected warnings while converting Nutanix CAPI resources to MAPI Machine")
		},

		// Base Case.
		Entry("passes with a base configuration", nutanixCAPI2MAPIMachineConversionInput{
			machine:          clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine:   createBaseNutanixMachine(),
			nutanixCluster:   createBaseNutanixCluster(),
			expectedErrors:   []string{},
			expectedWarnings: []string{},
		}),

		// Resource Identifier Errors.
		Entry("fails with empty name in image identifier", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.Image = &nutanixv1.NutanixResourceIdentifier{
					Type: nutanixv1.NutanixIdentifierName,
					Name: ptr.To(""),
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"Name cannot be empty for Name type identifier",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with nil name in image identifier", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.Image = &nutanixv1.NutanixResourceIdentifier{
					Type: nutanixv1.NutanixIdentifierName,
					Name: nil,
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"Name must be set for Name type identifier",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with empty UUID in cluster identifier", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.Cluster = nutanixv1.NutanixResourceIdentifier{
					Type: nutanixv1.NutanixIdentifierUUID,
					UUID: ptr.To(""),
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"UUID cannot be empty for UUID type identifier",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with nil UUID in cluster identifier", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.Cluster = nutanixv1.NutanixResourceIdentifier{
					Type: nutanixv1.NutanixIdentifierUUID,
					UUID: nil,
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"UUID must be set for UUID type identifier",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with empty identifier type", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.Image = &nutanixv1.NutanixResourceIdentifier{
					Type: "",
					Name: ptr.To("test-image"),
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"identifier type must be specified",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with invalid identifier type", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.Image = &nutanixv1.NutanixResourceIdentifier{
					Type: "invalid",
					Name: ptr.To("test-image"),
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"invalid identifier type, must be 'name' or 'uuid'",
			},
			expectedWarnings: []string{},
		}),

		// Boot Type Errors.
		Entry("fails with invalid boot type", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.BootType = "InvalidBootType"

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"invalid boot type, must be 'Legacy' or 'UEFI'",
			},
			expectedWarnings: []string{},
		}),

		// GPU Errors.
		Entry("fails with GPU DeviceID type but nil DeviceID", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.GPUs = []nutanixv1.NutanixGPU{
					{
						Type:     nutanixv1.NutanixGPUIdentifierDeviceID,
						DeviceID: nil,
					},
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"DeviceID must be set for DeviceID type GPU",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with GPU Name type but nil Name", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.GPUs = []nutanixv1.NutanixGPU{
					{
						Type: nutanixv1.NutanixGPUIdentifierName,
						Name: nil,
					},
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"Name must be set for Name type GPU",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with invalid GPU identifier type", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.GPUs = []nutanixv1.NutanixGPU{
					{
						Type: "InvalidGPUType",
					},
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"invalid GPU identifier type",
			},
			expectedWarnings: []string{},
		}),

		// Data Disk Errors.
		Entry("fails with invalid disk device type", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.DataDisks = []nutanixv1.NutanixMachineVMDisk{
					{
						DiskSize: resource.MustParse("100Gi"),
						DeviceProperties: &nutanixv1.NutanixMachineVMDiskDeviceProperties{
							DeviceType:  "InvalidDeviceType",
							AdapterType: nutanixv1.NutanixMachineDiskAdapterTypeSCSI,
						},
					},
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"DeviceType should be CDRom or Disk",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with invalid disk adapter type", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.DataDisks = []nutanixv1.NutanixMachineVMDisk{
					{
						DiskSize: resource.MustParse("100Gi"),
						DeviceProperties: &nutanixv1.NutanixMachineVMDiskDeviceProperties{
							DeviceType:  nutanixv1.NutanixMachineDiskDeviceTypeDisk,
							AdapterType: "InvalidAdapterType",
						},
					},
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"AdapterType can be SCSI, IDE, PCI, SATA or SPAPR",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with invalid storage disk mode", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.DataDisks = []nutanixv1.NutanixMachineVMDisk{
					{
						DiskSize: resource.MustParse("100Gi"),
						StorageConfig: &nutanixv1.NutanixMachineVMStorageConfig{
							DiskMode: "InvalidDiskMode",
						},
					},
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"DiskMode can be Standard and Flash",
			},
			expectedWarnings: []string{},
		}),
		Entry("fails with storage container name identifier (not supported)", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.DataDisks = []nutanixv1.NutanixMachineVMDisk{
					{
						DiskSize: resource.MustParse("100Gi"),
						StorageConfig: &nutanixv1.NutanixMachineVMStorageConfig{
							DiskMode: nutanixv1.NutanixMachineDiskModeStandard,
							StorageContainer: &nutanixv1.NutanixResourceIdentifier{
								Type: nutanixv1.NutanixIdentifierName,
								Name: ptr.To("storage-container"),
							},
						},
					},
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"storage resource identifiers only support UUID type",
			},
			expectedWarnings: []string{},
		}),

		// Warnings.
		Entry("warns with ImageLookup field present", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.ImageLookup = &nutanixv1.NutanixImageLookup{
					BaseOS: "rhcos",
					Format: ptr.To("capx-{{.BaseOS}}-{{.K8sVersion}}-*"),
				}

				return machine
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{},
			expectedWarnings: []string{
				"ImageLookup field is not supported in Machine API and will be ignored",
			},
		}),

		// Edge Cases.
		Entry("passes with empty boot type (defaults to Legacy)", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.BootType = ""

				return machine
			}(),
			nutanixCluster:   createBaseNutanixCluster(),
			expectedErrors:   []string{},
			expectedWarnings: []string{},
		}),
		Entry("passes with valid GPU configurations", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.GPUs = []nutanixv1.NutanixGPU{
					{
						Type:     nutanixv1.NutanixGPUIdentifierDeviceID,
						DeviceID: ptr.To(int64(1234)),
					},
					{
						Type: nutanixv1.NutanixGPUIdentifierName,
						Name: ptr.To("NVIDIA-GPU"),
					},
				}

				return machine
			}(),
			nutanixCluster:   createBaseNutanixCluster(),
			expectedErrors:   []string{},
			expectedWarnings: []string{},
		}),
		Entry("passes with valid categories", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.AdditionalCategories = []nutanixv1.NutanixCategoryIdentifier{
					{
						Key:   "Environment",
						Value: "Production",
					},
					{
						Key:   "Team",
						Value: "Platform",
					},
				}

				return machine
			}(),
			nutanixCluster:   createBaseNutanixCluster(),
			expectedErrors:   []string{},
			expectedWarnings: []string{},
		}),
		Entry("passes with project specified", nutanixCAPI2MAPIMachineConversionInput{
			machine: clusterv1resourcebuilder.Machine().Build(),
			nutanixMachine: func() *nutanixv1.NutanixMachine {
				machine := createBaseNutanixMachine()
				machine.Spec.Project = &nutanixv1.NutanixResourceIdentifier{
					Type: nutanixv1.NutanixIdentifierName,
					Name: ptr.To("openshift-project"),
				}

				return machine
			}(),
			nutanixCluster:   createBaseNutanixCluster(),
			expectedErrors:   []string{},
			expectedWarnings: []string{},
		}),
		Entry("passes with failure domain", nutanixCAPI2MAPIMachineConversionInput{
			machine:          clusterv1resourcebuilder.Machine().WithFailureDomain(ptr.To("fd-1")).Build(),
			nutanixMachine:   createBaseNutanixMachine(),
			nutanixCluster:   createBaseNutanixCluster(),
			expectedErrors:   []string{},
			expectedWarnings: []string{},
		}),
	)

	_ = DescribeTable("capi2mapi Nutanix convert CAPI MachineSet/InfraMachineTemplate/InfraCluster to MAPI MachineSet",
		func(in nutanixCAPI2MAPIMachinesetConversionInput) {
			_, warns, err := FromMachineSetAndNutanixMachineTemplateAndNutanixCluster(
				in.machineSet,
				in.nutanixMachineTemplate,
				in.nutanixCluster,
			).ToMachineSet()
			Expect(err).To(matchers.ConsistOfMatchErrorSubstrings(in.expectedErrors),
				"should match expected errors while converting Nutanix CAPI resources to MAPI MachineSet")
			Expect(warns).To(matchers.ConsistOfSubstrings(in.expectedWarnings),
				"should match expected warnings while converting Nutanix CAPI resources to MAPI MachineSet")
		},

		// Base Case.
		Entry("passes with a base configuration", nutanixCAPI2MAPIMachinesetConversionInput{
			machineSet:             clusterv1resourcebuilder.MachineSet().Build(),
			nutanixMachineTemplate: createBaseNutanixMachineTemplate(),
			nutanixCluster:         createBaseNutanixCluster(),
			expectedErrors:         []string{},
			expectedWarnings:       []string{},
		}),

		// Error Cases for MachineSet.
		Entry("fails with invalid resource identifier in template", nutanixCAPI2MAPIMachinesetConversionInput{
			machineSet: clusterv1resourcebuilder.MachineSet().Build(),
			nutanixMachineTemplate: func() *nutanixv1.NutanixMachineTemplate {
				template := createBaseNutanixMachineTemplate()
				template.Spec.Template.Spec.Image = &nutanixv1.NutanixResourceIdentifier{
					Type: nutanixv1.NutanixIdentifierName,
					Name: ptr.To(""),
				}

				return template
			}(),
			nutanixCluster: createBaseNutanixCluster(),
			expectedErrors: []string{
				"Name cannot be empty for Name type identifier",
			},
			expectedWarnings: []string{},
		}),
	)
})
