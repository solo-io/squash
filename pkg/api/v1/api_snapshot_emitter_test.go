// Code generated by solo-kit. DO NOT EDIT.

// +build solokit

package v1

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	"k8s.io/client-go/rest"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	// From https://github.com/kubernetes/client-go/blob/53c7adfd0294caa142d961e1f780f74081d5b15f/examples/out-of-cluster-client-configuration/main.go#L31
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("V1Emitter", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace1            string
		namespace2            string
		name1, name2          = "angela" + helpers.RandString(3), "bob" + helpers.RandString(3)
		cfg                   *rest.Config
		emitter               ApiEmitter
		debugAttachmentClient DebugAttachmentClient
	)

	BeforeEach(func() {
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
		var err error
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		err = setup.SetupKubeForTest(namespace1)
		Expect(err).NotTo(HaveOccurred())
		err = setup.SetupKubeForTest(namespace2)
		Expect(err).NotTo(HaveOccurred())
		// DebugAttachment Constructor
		debugAttachmentClientFactory := &factory.KubeResourceClientFactory{
			Crd:         DebugAttachmentCrd,
			Cfg:         cfg,
			SharedCache: kuberc.NewKubeCache(context.TODO()),
		}

		debugAttachmentClient, err = NewDebugAttachmentClient(debugAttachmentClientFactory)
		Expect(err).NotTo(HaveOccurred())
		emitter = NewApiEmitter(debugAttachmentClient)
	})
	AfterEach(func() {
		setup.TeardownKube(namespace1)
		setup.TeardownKube(namespace2)
	})
	It("tracks snapshots on changes to any resource", func() {
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		snapshots, errs, err := emitter.Snapshots([]string{namespace1, namespace2}, clients.WatchOpts{
			Ctx:         ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())

		var snap *ApiSnapshot

		/*
			DebugAttachment
		*/

		assertSnapshotDebugattachments := func(expectDebugattachments DebugAttachmentList, unexpectDebugattachments DebugAttachmentList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectDebugattachments {
						if _, err := snap.Debugattachments.List().Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectDebugattachments {
						if _, err := snap.Debugattachments.List().Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := debugAttachmentClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := debugAttachmentClient.List(namespace2, clients.ListOpts{})
					combined := DebugattachmentsByNamespace{
						namespace1: nsList1,
						namespace2: nsList2,
					}
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		debugAttachment1a, err := debugAttachmentClient.Write(NewDebugAttachment(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		debugAttachment1b, err := debugAttachmentClient.Write(NewDebugAttachment(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotDebugattachments(DebugAttachmentList{debugAttachment1a, debugAttachment1b}, nil)
		debugAttachment2a, err := debugAttachmentClient.Write(NewDebugAttachment(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		debugAttachment2b, err := debugAttachmentClient.Write(NewDebugAttachment(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotDebugattachments(DebugAttachmentList{debugAttachment1a, debugAttachment1b, debugAttachment2a, debugAttachment2b}, nil)

		err = debugAttachmentClient.Delete(debugAttachment2a.GetMetadata().Namespace, debugAttachment2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = debugAttachmentClient.Delete(debugAttachment2b.GetMetadata().Namespace, debugAttachment2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotDebugattachments(DebugAttachmentList{debugAttachment1a, debugAttachment1b}, DebugAttachmentList{debugAttachment2a, debugAttachment2b})

		err = debugAttachmentClient.Delete(debugAttachment1a.GetMetadata().Namespace, debugAttachment1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = debugAttachmentClient.Delete(debugAttachment1b.GetMetadata().Namespace, debugAttachment1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotDebugattachments(nil, DebugAttachmentList{debugAttachment1a, debugAttachment1b, debugAttachment2a, debugAttachment2b})
	})
	It("tracks snapshots on changes to any resource using AllNamespace", func() {
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		snapshots, errs, err := emitter.Snapshots([]string{""}, clients.WatchOpts{
			Ctx:         ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())

		var snap *ApiSnapshot

		/*
			DebugAttachment
		*/

		assertSnapshotDebugattachments := func(expectDebugattachments DebugAttachmentList, unexpectDebugattachments DebugAttachmentList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectDebugattachments {
						if _, err := snap.Debugattachments.List().Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectDebugattachments {
						if _, err := snap.Debugattachments.List().Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := debugAttachmentClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := debugAttachmentClient.List(namespace2, clients.ListOpts{})
					combined := DebugattachmentsByNamespace{
						namespace1: nsList1,
						namespace2: nsList2,
					}
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}
		debugAttachment1a, err := debugAttachmentClient.Write(NewDebugAttachment(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		debugAttachment1b, err := debugAttachmentClient.Write(NewDebugAttachment(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotDebugattachments(DebugAttachmentList{debugAttachment1a, debugAttachment1b}, nil)
		debugAttachment2a, err := debugAttachmentClient.Write(NewDebugAttachment(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		debugAttachment2b, err := debugAttachmentClient.Write(NewDebugAttachment(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotDebugattachments(DebugAttachmentList{debugAttachment1a, debugAttachment1b, debugAttachment2a, debugAttachment2b}, nil)

		err = debugAttachmentClient.Delete(debugAttachment2a.GetMetadata().Namespace, debugAttachment2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = debugAttachmentClient.Delete(debugAttachment2b.GetMetadata().Namespace, debugAttachment2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotDebugattachments(DebugAttachmentList{debugAttachment1a, debugAttachment1b}, DebugAttachmentList{debugAttachment2a, debugAttachment2b})

		err = debugAttachmentClient.Delete(debugAttachment1a.GetMetadata().Namespace, debugAttachment1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = debugAttachmentClient.Delete(debugAttachment1b.GetMetadata().Namespace, debugAttachment1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotDebugattachments(nil, DebugAttachmentList{debugAttachment1a, debugAttachment1b, debugAttachment2a, debugAttachment2b})
	})
})
