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
package capi2mapi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	nutanixv1 "github.com/nutanix-cloud-native/cluster-api-provider-nutanix/api/v1beta1"
	credentialTypes "github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/randfill"

	"github.com/openshift/cluster-capi-operator/pkg/conversion/capi2mapi"
	"github.com/openshift/cluster-capi-operator/pkg/conversion/mapi2capi"
	conversiontest "github.com/openshift/cluster-capi-operator/pkg/conversion/test/fuzz"
)

const (
	nutanixMachineKind  = "NutanixMachine"
	nutanixTemplateKind = "NutanixMachineTemplate"
)

var _ = Describe("Nutanix Fuzz (capi2mapi)", func() {
	infra := &configv1.Infrastructure{
		Spec: configv1.InfrastructureSpec{},
		Status: configv1.InfrastructureStatus{
			InfrastructureName: "sample-cluster-name",
		},
	}

	infraCluster := &nutanixv1.NutanixCluster{
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

	Context("NutanixMachine Conversion", func() {
		fromMachineAndNutanixMachineAndNutanixCluster := func(machine *clusterv1.Machine, infraMachine client.Object, infraCluster client.Object) capi2mapi.MachineAndInfrastructureMachine {
			nutanixMachine, ok := infraMachine.(*nutanixv1.NutanixMachine)
			Expect(ok).To(BeTrue(), "input infra machine should be of type %T, got %T", &nutanixv1.NutanixMachine{}, infraMachine)

			nutanixCluster, ok := infraCluster.(*nutanixv1.NutanixCluster)
			Expect(ok).To(BeTrue(), "input infra cluster should be of type %T, got %T", &nutanixv1.NutanixCluster{}, infraCluster)

			return capi2mapi.FromMachineAndNutanixMachineAndNutanixCluster(machine, nutanixMachine, nutanixCluster)
		}

		conversiontest.CAPI2MAPIMachineRoundTripFuzzTest(
			scheme,
			infra,
			infraCluster,
			&nutanixv1.NutanixMachine{},
			mapi2capi.FromNutanixMachineAndInfra,
			fromMachineAndNutanixMachineAndNutanixCluster,
			conversiontest.ObjectMetaFuzzerFuncs(capiNamespace),
			conversiontest.CAPIMachineFuzzerFuncs(nutanixProviderIDFuzzer, nutanixMachineKind, nutanixv1.GroupVersion.String(), infra.Status.InfrastructureName),
			nutanixMachineFuzzerFuncs,
		)
	})

	Context("NutanixMachineSet Conversion", func() {
		fromMachineSetAndNutanixMachineTemplateAndNutanixCluster := func(machineSet *clusterv1.MachineSet, infraMachineTemplate client.Object, infraCluster client.Object) capi2mapi.MachineSetAndMachineTemplate {
			nutanixMachineTemplate, ok := infraMachineTemplate.(*nutanixv1.NutanixMachineTemplate)
			Expect(ok).To(BeTrue(), "input infra machine template should be of type %T, got %T", &nutanixv1.NutanixMachineTemplate{}, infraMachineTemplate)

			nutanixCluster, ok := infraCluster.(*nutanixv1.NutanixCluster)
			Expect(ok).To(BeTrue(), "input infra cluster should be of type %T, got %T", &nutanixv1.NutanixCluster{}, infraCluster)

			return capi2mapi.FromMachineSetAndNutanixMachineTemplateAndNutanixCluster(machineSet, nutanixMachineTemplate, nutanixCluster)
		}

		conversiontest.CAPI2MAPIMachineSetRoundTripFuzzTest(
			scheme,
			infra,
			infraCluster,
			&nutanixv1.NutanixMachineTemplate{},
			mapi2capi.FromNutanixMachineSetAndInfra,
			fromMachineSetAndNutanixMachineTemplateAndNutanixCluster,
			conversiontest.ObjectMetaFuzzerFuncs(capiNamespace),
			conversiontest.CAPIMachineFuzzerFuncs(nutanixProviderIDFuzzer, nutanixTemplateKind, nutanixv1.GroupVersion.String(), infra.Status.InfrastructureName),
			conversiontest.CAPIMachineSetFuzzerFuncs(nutanixTemplateKind, nutanixv1.GroupVersion.String(), infra.Status.InfrastructureName),
			nutanixMachineFuzzerFuncs,
			nutanixMachineTemplateFuzzerFuncs,
		)
	})
})

func nutanixProviderIDFuzzer(c randfill.Continue) string {
	return "nutanix://" + uuid.NewString()
}

func nutanixMachineFuzzerFuncs(codecs runtimeserializer.CodecFactory) []any {
	return []any{
		func(identifier *nutanixv1.NutanixResourceIdentifier, c randfill.Continue) {
			// We require either a UUID or a Name, not both.
			switch c.Int31n(2) {
			case 0:
				identifier.Type = nutanixv1.NutanixIdentifierUUID
				identifier.UUID = ptr.To(uuid.NewString())
				identifier.Name = nil
			case 1:
				identifier.Type = nutanixv1.NutanixIdentifierName
				identifier.Name = ptr.To("test-resource-" + uuid.NewString()[:8])
				identifier.UUID = nil
			}
		},
		func(spec *nutanixv1.NutanixMachineSpec, c randfill.Continue) {
			c.FillNoCustom(spec)

			// Ensure valid CPU and memory configurations
			if spec.VCPUSockets == 0 {
				spec.VCPUSockets = 2
			}
			if spec.VCPUsPerSocket == 0 {
				spec.VCPUsPerSocket = 1
			}
			if spec.MemorySize.IsZero() {
				spec.MemorySize = resource.MustParse("4Gi")
			}
			if spec.SystemDiskSize.IsZero() {
				spec.SystemDiskSize = resource.MustParse("120Gi")
			}

			// Ensure valid boot type
			if spec.BootType == "" {
				spec.BootType = nutanixv1.NutanixBootTypeLegacy
			}

			// ImageLookup is not supported in MAPI, clear it
			spec.ImageLookup = nil
		},
		func(disk *nutanixv1.NutanixMachineVMDisk, c randfill.Continue) {
			c.FillNoCustom(disk)

			if disk.DiskSize.IsZero() {
				disk.DiskSize = resource.MustParse("100Gi")
			}
		},
		func(m *nutanixv1.NutanixMachine, c randfill.Continue) {
			c.FillNoCustom(m)

			// Ensure the type meta is set correctly.
			m.TypeMeta.APIVersion = nutanixv1.GroupVersion.String()
			m.TypeMeta.Kind = nutanixMachineKind
		},
	}
}

func nutanixMachineTemplateFuzzerFuncs(codecs runtimeserializer.CodecFactory) []any {
	return []any{
		func(m *nutanixv1.NutanixMachineTemplate, c randfill.Continue) {
			c.FillNoCustom(m)

			// Ensure the type meta is set correctly.
			m.TypeMeta.APIVersion = nutanixv1.GroupVersion.String()
			m.TypeMeta.Kind = nutanixTemplateKind
		},
	}
}
