package roundrobin

import (
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/client/loadbalancer"
	"testing"
)

func TestNew(t *testing.T) {
	si := serviceInfo(true)
	si2 := serviceInfo(true)

	lb := New([]skynet.ServiceInfo{si, si2}).(*LoadBalancer)

	if len(lb.instances) != 2 {
		t.Fatal("Failed to update instances", len(lb.instances))
	}

	if lb.instanceList.Len() != 2 {
		t.Fatal("Failed to update list", lb.instanceList.Len())
	}
}

func TestAdd(t *testing.T) {
	lb := New([]skynet.ServiceInfo{}).(*LoadBalancer)

	si := serviceInfo(true)
	si2 := serviceInfo(true)
	si3 := serviceInfo(true)

	lb.AddInstance(si)
	lb.AddInstance(si2)
	lb.AddInstance(si3)

	if len(lb.instances) != 3 {
		t.Fatal("Failed to update instances")
	}

	if lb.instanceList.Len() != 3 {
		t.Fatal("Failed to update list")
	}

	// Ensure items are added in correct order
	if lb.instanceList.Front().Value.(skynet.ServiceInfo).UUID != si.UUID {
		t.Fatal("Failed to add instances in the correct order")
	}

	if lb.instanceList.Back().Value.(skynet.ServiceInfo).UUID != si3.UUID {
		t.Fatal("Failed to add instances in the correct order")
	}
}

func TestAddIgnoresDuplicates(t *testing.T) {
	lb := New([]skynet.ServiceInfo{}).(*LoadBalancer)

	si := serviceInfo(true)

	lb.AddInstance(si)
	lb.AddInstance(si)

	if len(lb.instances) != 1 {
		t.Fatal("Add did not ignore duplicates", len(lb.instances))
	}

	if lb.instanceList.Len() != 1 {
		t.Fatal("Add did not ignore duplicates", lb.instanceList.Len())
	}
}

func TestAddUpdatesDuplicate(t *testing.T) {
	lb := New([]skynet.ServiceInfo{}).(*LoadBalancer)

	si := serviceInfo(true)

	lb.AddInstance(si)
	si.Name = "Foo"
	lb.AddInstance(si)

	if len(lb.instances) != 1 {
		t.Fatal("Add did not ignore duplicates", len(lb.instances))
	}

	if lb.instanceList.Len() != 1 {
		t.Fatal("Add did not ignore duplicates", lb.instanceList.Len())
	}

	if lb.instances[si.UUID].Value.(skynet.ServiceInfo).Name != "Foo" {
		t.Fatal("Existing instance was not updated")
	}
}

func TestAddDoesNotAddUnregisteredToList(t *testing.T) {
	lb := New([]skynet.ServiceInfo{}).(*LoadBalancer)

	si := serviceInfo(true)
	si2 := serviceInfo(false)

	lb.AddInstance(si)
	lb.AddInstance(si2)

	if len(lb.instances) != 2 {
		t.Fatal("Failed to update instances", len(lb.instances))
	}

	if lb.instanceList.Len() != 1 {
		t.Fatal("Unregistered instances should not be part of list", lb.instanceList.Len())
	}
}

func TestUpdate(t *testing.T) {
	si := serviceInfo(true)
	si2 := serviceInfo(true)
	si3 := serviceInfo(true)

	lb := New([]skynet.ServiceInfo{si, si2, si3}).(*LoadBalancer)

	si2.Name = "Foo"
	lb.UpdateInstance(si2)

	if lb.instances[si2.UUID].Value.(skynet.ServiceInfo).Name != si2.Name {
		t.Fatal("Existing instance was not updated")
	}
}
func TestUpdateAddsInstanceIfItDoesntExist(t *testing.T) {
	si := serviceInfo(true)

	lb := New([]skynet.ServiceInfo{}).(*LoadBalancer)

	lb.UpdateInstance(si)

	if len(lb.instances) != 1 {
		t.Fatal("Failed to update instances")
	}

	if lb.instanceList.Len() != 1 {
		t.Fatal("Failed to update list")
	}
}

func TestUpdateRemovesUnregisteredFromList(t *testing.T) {
	si := serviceInfo(true)

	lb := New([]skynet.ServiceInfo{si}).(*LoadBalancer)

	si.Registered = false
	lb.UpdateInstance(si)

	if len(lb.instances) != 1 {
		t.Fatal("Unregistered instances should still be tracked")
	}

	if lb.instanceList.Len() != 0 {
		t.Fatal("Unregistered instances should not be in list")
	}
}

func TestRemove(t *testing.T) {
	si := serviceInfo(true)
	si2 := serviceInfo(true)
	si3 := serviceInfo(true)
	si4 := serviceInfo(true)

	lb := New([]skynet.ServiceInfo{si, si2, si3, si4}).(*LoadBalancer)

	lb.RemoveInstance(si4)

	if len(lb.instances) != 3 {
		t.Fatal("Failed to update instances")
	}

	if lb.instanceList.Len() != 3 {
		t.Fatal("Failed to update list")
	}

	// Ensure items are added in correct order
	if lb.instanceList.Front().Value.(skynet.ServiceInfo).UUID != si.UUID {
		t.Fatal("Failed to add instances in the correct order")
	}

	if lb.instanceList.Back().Value.(skynet.ServiceInfo).UUID != si3.UUID {
		t.Fatal("Failed to add instances in the correct order")
	}
}

func TestChooseReturnsErrorWhenEmpty(t *testing.T) {
	lb := New([]skynet.ServiceInfo{}).(*LoadBalancer)

	_, err := lb.Choose()

	if err != loadbalancer.NoInstances {
		t.Fatal("LoadBalancer should fail if no instances exist")
	}
}

func TestChoose(t *testing.T) {
	instances := []skynet.ServiceInfo{serviceInfo(true), serviceInfo(true), serviceInfo(true), serviceInfo(true)}

	lb := New(instances).(*LoadBalancer)

	// Check order
	for i := 0; i <= 3; i++ {
		s, err := lb.Choose()

		if err != nil || s.UUID != instances[i].UUID {
			t.Fatal("LoadBalancer did not properly iterate over instances")
		}
	}

	// Ensure Choose loops around when it hits the end
	s, err := lb.Choose()
	if err != nil || s.UUID != instances[0].UUID {
		t.Fatal("LoadBalancer did not properly iterate over instances")
	}
}

func serviceInfo(registered bool) skynet.ServiceInfo {
	si := skynet.NewServiceInfo(nil)
	si.Registered = registered

	return si
}
