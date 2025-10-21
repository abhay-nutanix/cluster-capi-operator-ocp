package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	nutanixv1 "github.com/nutanix-cloud-native/cluster-api-provider-nutanix/api/v1beta1"
	configv1 "github.com/openshift/api/config/v1"
	mapiv1 "github.com/openshift/api/machine/v1"
	mapiv1beta1 "github.com/openshift/api/machine/v1beta1"
	framework "github.com/openshift/cluster-capi-operator/e2e/framework"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	nutanixMachineTemplateName = "nutanix-e2e-template"
)

var _ = Describe("Cluster API Nutanix MachineSet", Ordered, func() {
	var (
		nutanixMachineTemplate *nutanixv1.NutanixMachineTemplate
		machineSet             *clusterv1.MachineSet
		mapiSpec               *mapiv1.NutanixMachineProviderConfig
	)

	BeforeAll(func() {
		if platform != configv1.NutanixPlatformType {
			Skip("Skipping Nutanix E2E tests")
		}
		mapiSpec = getNutanixMAPIProviderSpec(cl)
	})

	AfterEach(func() {
		if platform != configv1.NutanixPlatformType {
			Skip("Skipping Nutanix E2E tests")
		}
		framework.DeleteMachineSets(ctx, cl, machineSet)
		framework.WaitForMachineSetsDeleted(cl, machineSet)
		framework.DeleteObjects(ctx, cl, nutanixMachineTemplate)
	})

	It("should be able to run a machine", func() {
		nutanixMachineTemplate = createNutanixMachineTemplate(ctx, cl, mapiSpec)

		machineSet = framework.CreateMachineSet(ctx, cl, framework.NewMachineSetParams(
			"nutanix-machineset",
			clusterName,
			"",
			1,
			corev1.ObjectReference{
				Kind:       "NutanixMachineTemplate",
				APIVersion: infraAPIVersion,
				Name:       nutanixMachineTemplateName,
			},
			"worker-user-data",
		))

		framework.WaitForMachineSet(cl, machineSet.Name, machineSet.Namespace)
	})
})

func getNutanixMAPIProviderSpec(cl client.Client) *mapiv1.NutanixMachineProviderConfig {
	machineSetList := &mapiv1beta1.MachineSetList{}
	Expect(cl.List(ctx, machineSetList, client.InNamespace(framework.MAPINamespace))).To(Succeed())
	Expect(machineSetList.Items).ToNot(HaveLen(0))

	machineSet := machineSetList.Items[0]
	Expect(machineSet.Spec.Template.Spec.ProviderSpec.Value).ToNot(BeNil())

	providerSpec := &mapiv1.NutanixMachineProviderConfig{}
	Expect(yaml.Unmarshal(machineSet.Spec.Template.Spec.ProviderSpec.Value.Raw, providerSpec)).To(Succeed())

	return providerSpec
}

func createNutanixMachineTemplate(ctx context.Context, cl client.Client, m *mapiv1.NutanixMachineProviderConfig) *nutanixv1.NutanixMachineTemplate {
	By("Creating Nutanix machine template")

	Expect(m).ToNot(BeNil())
	Expect(m.Cluster.Type).To(Equal(mapiv1.NutanixIdentifierUUID))
	Expect(m.Cluster.UUID).ToNot(BeNil())
	Expect(m.Subnets).ToNot(BeEmpty())
	Expect(m.SystemDiskSize.Value()).To(BeNumerically(">", 0))
	Expect(m.MemorySize.Value()).To(BeNumerically(">", 0))
	Expect(m.VCPUSockets).To(BeNumerically(">=", 1))
	Expect(m.VCPUsPerSocket).To(BeNumerically(">=", 1))

	capx := &nutanixv1.NutanixMachineTemplate{
		Spec: nutanixv1.NutanixMachineTemplateSpec{
			Template: nutanixv1.NutanixMachineTemplateResource{
				Spec: nutanixv1.NutanixMachineSpec{
					VCPUSockets:    m.VCPUSockets,
					VCPUsPerSocket: m.VCPUsPerSocket,
					MemorySize:     m.MemorySize,
					SystemDiskSize: m.SystemDiskSize,
					Cluster: nutanixv1.NutanixResourceIdentifier{
						Type: nutanixv1.NutanixIdentifierUUID,
						UUID: m.Cluster.UUID,
					},
					Subnets:              convertSubnets(m.Subnets),
					BootType:             convertBootType(m.BootType),
					Image:                convertImage(m.Image),
					Project:              convertResource(m.Project),
					AdditionalCategories: convertCategories(m.Categories),
				},
			},
		},
	}
	capx.TypeMeta.APIVersion = nutanixv1.GroupVersion.String()
	capx.TypeMeta.Kind = "NutanixMachineTemplate"
	capx.ObjectMeta.Name = nutanixMachineTemplateName
	capx.ObjectMeta.Namespace = framework.CAPINamespace

	Expect(cl.Create(ctx, capx)).To(Succeed())
	return capx
}

func convertSubnets(in []mapiv1.NutanixResourceIdentifier) []nutanixv1.NutanixResourceIdentifier {
	out := make([]nutanixv1.NutanixResourceIdentifier, 0, len(in))
	for _, s := range in {
		if s.Type == mapiv1.NutanixIdentifierUUID && s.UUID != nil {
			out = append(out, nutanixv1.NutanixResourceIdentifier{Type: nutanixv1.NutanixIdentifierUUID, UUID: s.UUID})
		}
	}
	return out
}

func convertBootType(in mapiv1.NutanixBootType) nutanixv1.NutanixBootType {
	switch in {
	case mapiv1.NutanixUEFIBoot:
		return nutanixv1.NutanixBootTypeUEFI
	default:
		return nutanixv1.NutanixBootTypeLegacy
	}
}

func convertImage(in mapiv1.NutanixResourceIdentifier) *nutanixv1.NutanixResourceIdentifier {
	if in.Type == "" {
		return nil
	}
	out := &nutanixv1.NutanixResourceIdentifier{}
	switch in.Type {
	case mapiv1.NutanixIdentifierName:
		out.Type = nutanixv1.NutanixIdentifierName
		out.Name = in.Name
	case mapiv1.NutanixIdentifierUUID:
		out.Type = nutanixv1.NutanixIdentifierUUID
		out.UUID = in.UUID
	}
	return out
}

func convertResource(in mapiv1.NutanixResourceIdentifier) *nutanixv1.NutanixResourceIdentifier {
	if in.Type == "" {
		return nil
	}
	return &nutanixv1.NutanixResourceIdentifier{Type: nutanixv1.NutanixIdentifierUUID, UUID: in.UUID}
}

func convertCategories(in []mapiv1.NutanixCategory) []nutanixv1.NutanixCategoryIdentifier {
	if len(in) == 0 {
		return nil
	}
	out := make([]nutanixv1.NutanixCategoryIdentifier, 0, len(in))
	for _, c := range in {
		out = append(out, nutanixv1.NutanixCategoryIdentifier{Key: c.Key, Value: c.Value})
	}
	return out
}
