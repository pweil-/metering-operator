package fake

import (
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeScheduledReports implements ScheduledReportInterface
type FakeScheduledReports struct {
	Fake *FakeChargebackV1alpha1
	ns   string
}

var scheduledreportsResource = schema.GroupVersionResource{Group: "chargeback.coreos.com", Version: "v1alpha1", Resource: "scheduledreports"}

var scheduledreportsKind = schema.GroupVersionKind{Group: "chargeback.coreos.com", Version: "v1alpha1", Kind: "ScheduledReport"}

// Get takes name of the scheduledReport, and returns the corresponding scheduledReport object, and an error if there is any.
func (c *FakeScheduledReports) Get(name string, options v1.GetOptions) (result *v1alpha1.ScheduledReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(scheduledreportsResource, c.ns, name), &v1alpha1.ScheduledReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ScheduledReport), err
}

// List takes label and field selectors, and returns the list of ScheduledReports that match those selectors.
func (c *FakeScheduledReports) List(opts v1.ListOptions) (result *v1alpha1.ScheduledReportList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(scheduledreportsResource, scheduledreportsKind, c.ns, opts), &v1alpha1.ScheduledReportList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ScheduledReportList{}
	for _, item := range obj.(*v1alpha1.ScheduledReportList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested scheduledReports.
func (c *FakeScheduledReports) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(scheduledreportsResource, c.ns, opts))

}

// Create takes the representation of a scheduledReport and creates it.  Returns the server's representation of the scheduledReport, and an error, if there is any.
func (c *FakeScheduledReports) Create(scheduledReport *v1alpha1.ScheduledReport) (result *v1alpha1.ScheduledReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(scheduledreportsResource, c.ns, scheduledReport), &v1alpha1.ScheduledReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ScheduledReport), err
}

// Update takes the representation of a scheduledReport and updates it. Returns the server's representation of the scheduledReport, and an error, if there is any.
func (c *FakeScheduledReports) Update(scheduledReport *v1alpha1.ScheduledReport) (result *v1alpha1.ScheduledReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(scheduledreportsResource, c.ns, scheduledReport), &v1alpha1.ScheduledReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ScheduledReport), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeScheduledReports) UpdateStatus(scheduledReport *v1alpha1.ScheduledReport) (*v1alpha1.ScheduledReport, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(scheduledreportsResource, "status", c.ns, scheduledReport), &v1alpha1.ScheduledReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ScheduledReport), err
}

// Delete takes name of the scheduledReport and deletes it. Returns an error if one occurs.
func (c *FakeScheduledReports) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(scheduledreportsResource, c.ns, name), &v1alpha1.ScheduledReport{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeScheduledReports) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(scheduledreportsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ScheduledReportList{})
	return err
}

// Patch applies the patch and returns the patched scheduledReport.
func (c *FakeScheduledReports) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ScheduledReport, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(scheduledreportsResource, c.ns, name, data, subresources...), &v1alpha1.ScheduledReport{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ScheduledReport), err
}
