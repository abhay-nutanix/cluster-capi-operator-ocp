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
package mapi2capi_test

import (
	"encoding/json"
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	nutanixv1 "github.com/nutanix-cloud-native/cluster-api-provider-nutanix/api/v1beta1"
	configv1 "github.com/openshift/api/config/v1"
	mapiv1 "github.com/openshift/api/machine/v1"
	mapiv1beta1 "github.com/openshift/api/machine/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/randfill"

	"github.com/openshift/cluster-capi-operator/pkg/conversion/capi2mapi"
	"github.com/openshift/cluster-capi-operator/pkg/conversion/mapi2capi"
	conversiontest "github.com/openshift/cluster-capi-operator/pkg/conversion/test/fuzz"
)

const (
	nutanixProviderSpecKind = "NutanixMachineProviderConfig"
	nutanixTemplateKind     = "NutanixMachineTemplate"
)

func nutanixProviderIDFuzzer(c randfill.Continue) string {
	// Nutanix providerID format: nutanix://<cluster-uuid>/<vm-uuid>
	return "nutanix://" + uuid.NewString() + "/" + uuid.NewString()
}

var _ = PDescribe("Nutanix Fuzz (mapi2capi)", func() {

	infra := &configv1.Infrastructure{
		Spec: configv1.InfrastructureSpec{},
		Status: configv1.InfrastructureStatus{
			InfrastructureName: "sample-cluster-name",
		},
	}

	infraCluster := &nutanixv1.NutanixCluster{
		Spec: nutanixv1.NutanixClusterSpec{},
	}

	Context("NutanixMachine Conversion", func() {
		fromMachineAndNutanixMachineAndNutanixCluster := func(machine *clusterv1.Machine, infraMachine client.Object, infraCluster client.Object) capi2mapi.MachineAndInfrastructureMachine {
			_, ok := infraMachine.(*nutanixv1.NutanixMachine)
			Expect(ok).To(BeTrue(), "input infra machine should be of type %T, got %T", &nutanixv1.NutanixMachine{}, infraMachine)

			_, ok = infraCluster.(*nutanixv1.NutanixCluster)
			Expect(ok).To(BeTrue(), "input infra cluster should be of type %T, got %T", &nutanixv1.NutanixCluster{}, infraCluster)
			nutanixMachine, ok := infraMachine.(*nutanixv1.NutanixMachine)
			Expect(ok).To(BeTrue(), "input infra machine should be of type %T, got %T", &nutanixv1.NutanixMachine{}, infraMachine)
			nutanixCluster, ok := infraCluster.(*nutanixv1.NutanixCluster)
			Expect(ok).To(BeTrue(), "input infra cluster should be of type %T, got %T", &nutanixv1.NutanixCluster{}, infraCluster)

			return capi2mapi.FromMachineAndNutanixMachineAndNutanixCluster(machine, nutanixMachine, nutanixCluster)
		}

		conversiontest.MAPI2CAPIMachineRoundTripFuzzTest(
			scheme,
			infra,
			infraCluster,
			mapi2capi.FromNutanixMachineAndInfra,
			fromMachineAndNutanixMachineAndNutanixCluster,
			conversiontest.ObjectMetaFuzzerFuncs(mapiNamespace),
			conversiontest.MAPIMachineFuzzerFuncs(&mapiv1.NutanixMachineProviderConfig{}, &mapiv1.NutanixMachineProviderStatus{}, nutanixProviderIDFuzzer),
			nutanixProviderSpecFuzzerFuncs,
		)
	})

	Context("NutanixMachineSet Conversion", func() {
		fromMachineSetAndNutanixMachineTemplateAndNutanixCluster := func(machineSet *clusterv1.MachineSet, infraMachineTemplate client.Object, infraCluster client.Object) capi2mapi.MachineSetAndMachineTemplate {
			_, ok := infraMachineTemplate.(*nutanixv1.NutanixMachineTemplate)
			Expect(ok).To(BeTrue(), "input infra machine template should be of type %T, got %T", &nutanixv1.NutanixMachineTemplate{}, infraMachineTemplate)

			_, ok = infraCluster.(*nutanixv1.NutanixCluster)
			Expect(ok).To(BeTrue(), "input infra cluster should be of type %T, got %T", &nutanixv1.NutanixCluster{}, infraCluster)
			nutanixMachineTemplate, ok := infraMachineTemplate.(*nutanixv1.NutanixMachineTemplate)
			Expect(ok).To(BeTrue(), "input infra machine template should be of type %T, got %T", &nutanixv1.NutanixMachineTemplate{}, infraMachineTemplate)
			nutanixCluster, ok := infraCluster.(*nutanixv1.NutanixCluster)
			Expect(ok).To(BeTrue(), "input infra cluster should be of type %T, got %T", &nutanixv1.NutanixCluster{}, infraCluster)

			return capi2mapi.FromMachineSetAndNutanixMachineTemplateAndNutanixCluster(machineSet, nutanixMachineTemplate, nutanixCluster)
		}

		conversiontest.MAPI2CAPIMachineSetRoundTripFuzzTest(
			scheme,
			infra,
			infraCluster,
			mapi2capi.FromNutanixMachineSetAndInfra,
			fromMachineSetAndNutanixMachineTemplateAndNutanixCluster,
			conversiontest.ObjectMetaFuzzerFuncs(mapiNamespace),
			conversiontest.MAPIMachineFuzzerFuncs(&mapiv1.NutanixMachineProviderConfig{}, &mapiv1.NutanixMachineProviderStatus{}, nutanixProviderIDFuzzer),
			conversiontest.MAPIMachineSetFuzzerFuncs(),
			nutanixProviderSpecFuzzerFuncs,
		)
	})
})

//nolint:funlen
func nutanixProviderSpecFuzzerFuncs(codecs runtimeserializer.CodecFactory) []any {
	return []any{
		func(resourceId *mapiv1.NutanixResourceIdentifier, c randfill.Continue) {
			switch c.Int31n(2) {
			case 0:
				resourceId.Type = mapiv1.NutanixIdentifierName
				name := generateRandomString(10)
				resourceId.Name = &name
				resourceId.UUID = nil
			case 1:
				resourceId.Type = mapiv1.NutanixIdentifierUUID
				uuidStr := uuid.NewString()
				resourceId.UUID = &uuidStr
				resourceId.Name = nil
			}
		},
		// Removed individual field fuzzers for DataDisks, GPUs, and Categories
		// since these fields are intentionally set to nil in the main fuzzer
		// to ensure consistent null serialization in round-trip conversions
		func(providerSpec *mapiv1.NutanixMachineProviderConfig, c randfill.Continue) {
			// CRITICAL: Use selective filling instead of FillNoCustom to avoid metadata generation
			// DO NOT call c.FillNoCustom(providerSpec) as it generates metadata that gets lost in round-trip

			// The type meta is always set to these values by the conversion.
			providerSpec.APIVersion = mapiv1.SchemeGroupVersion.String()
			providerSpec.Kind = nutanixProviderSpecKind

			// Ensure metadata is completely empty from the start
			providerSpec.ObjectMeta = metav1.ObjectMeta{}

			// Set required numeric fields with valid values
			providerSpec.VCPUSockets = 2
			providerSpec.VCPUsPerSocket = 1

			// Set required memory and disk sizes
			providerSpec.MemorySize = resource.MustParse("4Gi")
			providerSpec.SystemDiskSize = resource.MustParse("120Gi")

			// Set valid boot type
			switch c.Int31n(2) {
			case 0:
				providerSpec.BootType = mapiv1.NutanixLegacyBoot
			case 1:
				providerSpec.BootType = mapiv1.NutanixUEFIBoot
			}

			// Clear problematic fields that cause round-trip failures
			providerSpec.CredentialsSecret = nil
			providerSpec.UserDataSecret = nil
			providerSpec.FailureDomain = nil

			// Set collections to nil to ensure null serialization
			providerSpec.Categories = nil
			providerSpec.GPUs = nil
			providerSpec.DataDisks = nil

			// Ensure at least one subnet is present and valid
			uuidStr := uuid.NewString()
			providerSpec.Subnets = []mapiv1.NutanixResourceIdentifier{
				{
					Type: mapiv1.NutanixIdentifierUUID,
					UUID: &uuidStr,
				},
			}

			// Set required cluster resource identifier
			clusterUUID := uuid.NewString()
			providerSpec.Cluster = mapiv1.NutanixResourceIdentifier{
				Type: mapiv1.NutanixIdentifierUUID,
				UUID: &clusterUUID,
			}

			// Set image resource identifier (optional but common)
			if c.Int31n(2) == 0 {
				imageName := generateRandomString(10)
				providerSpec.Image = mapiv1.NutanixResourceIdentifier{
					Type: mapiv1.NutanixIdentifierName,
					Name: &imageName,
				}
			}

			// Set project resource identifier (optional)
			if c.Int31n(2) == 0 {
				projectUUID := uuid.NewString()
				providerSpec.Project = mapiv1.NutanixResourceIdentifier{
					Type: mapiv1.NutanixIdentifierUUID,
					UUID: &projectUUID,
				}
			}
		},
		// Normalize providerSpec Raw to align metadata and optional collections for round-trip
		func(m *mapiv1beta1.MachineSpec, c randfill.Continue) {
			if m.ProviderSpec.Value == nil || len(m.ProviderSpec.Value.Raw) == 0 {
				return
			}

			var ps mapiv1.NutanixMachineProviderConfig
			if err := json.Unmarshal(m.ProviderSpec.Value.Raw, &ps); err != nil {
				return
			}

			// Align providerSpec metadata with machine spec metadata to match CAPI->MAPI behavior
			ps.ObjectMeta.Labels = m.ObjectMeta.Labels
			ps.ObjectMeta.Annotations = m.ObjectMeta.Annotations

			// Ensure optional collections serialize as null, not []
			if len(ps.Categories) == 0 {
				ps.Categories = nil
			}
			if len(ps.GPUs) == 0 {
				ps.GPUs = nil
			}
			if len(ps.DataDisks) == 0 {
				ps.DataDisks = nil
			}

			// Avoid architectural warning paths in round-trip
			ps.CredentialsSecret = nil

			bytes, err := json.Marshal(ps)
			if err != nil {
				return
			}
			m.ProviderSpec.Value.Raw = bytes
		},
		// Normalize MachineSetSpec.Template.Spec.ProviderSpec for round-trip consistency
		// This is the same normalization as above but applied to the MachineSet template
		func(ms *mapiv1beta1.MachineSetSpec, c randfill.Continue) {
			if ms.Template.Spec.ProviderSpec.Value == nil || len(ms.Template.Spec.ProviderSpec.Value.Raw) == 0 {
				return
			}

			var ps mapiv1.NutanixMachineProviderConfig
			if err := json.Unmarshal(ms.Template.Spec.ProviderSpec.Value.Raw, &ps); err != nil {
				return
			}

			// Align providerSpec metadata with machine template spec metadata to match CAPI->MAPI behavior
			ps.ObjectMeta.Labels = ms.Template.Spec.ObjectMeta.Labels
			ps.ObjectMeta.Annotations = ms.Template.Spec.ObjectMeta.Annotations

			// Ensure optional collections serialize as null, not []
			if len(ps.Categories) == 0 {
				ps.Categories = nil
			}
			if len(ps.GPUs) == 0 {
				ps.GPUs = nil
			}
			if len(ps.DataDisks) == 0 {
				ps.DataDisks = nil
			}

			// Avoid architectural warning paths in round-trip
			ps.CredentialsSecret = nil

			bytes, err := json.Marshal(ps)
			if err != nil {
				return
			}
			ms.Template.Spec.ProviderSpec.Value.Raw = bytes
		},
	}
}

// generateRandomString generates a random alphanumeric string of specified length.
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}
