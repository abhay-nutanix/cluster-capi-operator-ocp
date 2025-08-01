package framework

import (
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

// GetMachines gets a list of machines from the default cluster API namespace.
// Optionaly, labels may be used to constrain listed machinesets.
func GetMachines(cl client.Client, selectors ...*metav1.LabelSelector) ([]*clusterv1.Machine, error) {
	machineList := &clusterv1.MachineList{}

	listOpts := append([]client.ListOption{},
		client.InNamespace(CAPINamespace),
	)

	for _, selector := range selectors {
		s, err := metav1.LabelSelectorAsSelector(selector)
		if err != nil {
			return nil, err
		}

		listOpts = append(listOpts,
			client.MatchingLabelsSelector{Selector: s},
		)
	}

	if err := cl.List(ctx, machineList, listOpts...); err != nil {
		return nil, fmt.Errorf("error querying api for machineList object: %w", err)
	}

	var machines []*clusterv1.Machine

	for i := range machineList.Items {
		machines = append(machines, &machineList.Items[i])
	}

	return machines, nil
}

// FilterRunningMachines returns a slice of only those Machines in the input
// that are in the "Running" phase.
func FilterRunningMachines(machines []*clusterv1.Machine) []*clusterv1.Machine {
	var result []*clusterv1.Machine

	for i, m := range machines {
		if m.Status.Phase == string(clusterv1.MachinePhaseRunning) {
			result = append(result, machines[i])
		}
	}

	return result
}

// GetMachine get a machine by its name.
func GetMachine(cl client.Client, name string, namespace string) (*clusterv1.Machine, error) {
	machine := &clusterv1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	Eventually(komega.Get(machine), time.Minute, RetryShort).Should(Succeed(), "Failed to get machine %s/%s.", machine.Namespace, machine.Name)

	return machine, nil
}

// DeleteMachines deletes the specified machines and returns an error on failure.
func DeleteMachines(cl client.Client, namespace string, machines ...*clusterv1.Machine) error {
	// 1. delete all machines
	for _, machine := range machines {
		if machine == nil {
			continue
		}
		Eventually(func() error {
			return cl.Delete(ctx, machine)
		}, time.Minute, RetryShort).Should(SatisfyAny(
			Succeed(),
			WithTransform(apierrors.IsNotFound, BeTrue()),
		), "Delete machine %s/%s should succeed, or machine should not be found.",
			machine.Namespace, machine.Name)
	}

	// 2. waiting for all machines to be deleted
	machineNames := []string{}
	for _, machine := range machines {
		machineNames = append(machineNames, machine.Name)
	}

	machineList := &clusterv1.MachineList{}
	Eventually(komega.ObjectList(machineList, client.InNamespace(namespace)), WaitLong, RetryMedium).Should(
		WithTransform(func(list *clusterv1.MachineList) []clusterv1.Machine {
			return list.Items
		}, Not(ContainElements(
			HaveField("ObjectMeta.Name", BeElementOf(machineNames)),
		))),
	)

	return nil
}
