package kubeconfig

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/cluster-api-actuator-pkg/testutils"
	"github.com/openshift/cluster-capi-operator/pkg/controllers"
)

var _ = Describe("Generate kubeconfig", func() {
	var options *kubeconfigOptions
	testBase64Text := "dGVzdA=="

	BeforeEach(func() {
		By("Creating the cluster-api namepsace")
		managedNamespace := &corev1.Namespace{}
		managedNamespace.SetName(controllers.DefaultManagedNamespace)
		Expect(cl.Create(ctx, managedNamespace)).To(Succeed())

		options = &kubeconfigOptions{
			token:            []byte(testBase64Text),
			caCert:           []byte(testBase64Text),
			apiServerEnpoint: "https://example.com",
			clusterName:      "test",
		}
	})

	AfterEach(func() {
		testutils.CleanupResources(Default, ctx, testEnv.Config, cl, controllers.DefaultManagedNamespace, &corev1.Secret{})
	})

	It("should generate kubeconfig", func() {
		kubeconfig, err := generateKubeconfig(*options)
		Expect(err).NotTo(HaveOccurred())
		Expect(kubeconfig).NotTo(BeNil())

		Expect(kubeconfig.Clusters).To(HaveKey(options.clusterName))
		Expect(kubeconfig.Clusters[options.clusterName].Server).To(Equal(options.apiServerEnpoint))
		Expect(kubeconfig.Clusters[options.clusterName].CertificateAuthorityData).To(Equal(options.caCert))

		Expect(kubeconfig.Contexts).To(HaveKey(options.clusterName))
		Expect(kubeconfig.Contexts[options.clusterName].Cluster).To(Equal(options.clusterName))
		Expect(kubeconfig.Contexts[options.clusterName].AuthInfo).To(Equal("cluster-capi-operator"))
		Expect(kubeconfig.Contexts[options.clusterName].Namespace).To(Equal(controllers.DefaultManagedNamespace))

		Expect(kubeconfig.AuthInfos).To(HaveKey("cluster-capi-operator"))
		Expect(kubeconfig.AuthInfos["cluster-capi-operator"].Token).To(Equal(testBase64Text))
	})

	It("should fail with empty token", func() {
		options.token = nil
		kubeconfig, err := generateKubeconfig(*options)
		Expect(err).To((HaveOccurred()))
		Expect(kubeconfig).To(BeNil())
	})

	It("should fail with empty ca cert", func() {
		options.caCert = nil
		kubeconfig, err := generateKubeconfig(*options)
		Expect(err).To((HaveOccurred()))
		Expect(kubeconfig).To(BeNil())
	})

	It("should fail with empty api server endpoint", func() {
		options.apiServerEnpoint = ""
		kubeconfig, err := generateKubeconfig(*options)
		Expect(err).To((HaveOccurred()))
		Expect(kubeconfig).To(BeNil())
	})

	It("should fail with empty cluster name", func() {
		options.clusterName = ""
		kubeconfig, err := generateKubeconfig(*options)
		Expect(err).To((HaveOccurred()))
		Expect(kubeconfig).To(BeNil())
	})
})
