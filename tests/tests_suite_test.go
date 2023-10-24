package tests_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")
}

// var _ = Describe("Status Test", func() {

// 	BeforeEach(func() {

// 	})

// 	Describe("Get Release", func() {
// 		Context("Online Update", func() {
// 			var release *codegen.Release
// 			var err error

// 			tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-1")
// 			Expect(err).ToNot(HaveOccurred())
// 			sysRoot := tmpDir
// 			fixtures.SetLocalRelease(sysRoot, "v0.4.4")

// 			statusService := &service.StatusService{
// 				ImplementService: &service.TestService{
// 					InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
// 				},
// 				SysRoot: sysRoot,
// 			}
// 			fmt.Println("init")

// 			It("should state is null", func() {
// 				fmt.Println("1")

// 				value, msg := service.GetStatus()
// 				Expect(value.Status).To(Equal(codegen.Idle))
// 				Expect(msg).To(Equal(""))
// 			})

// 			ctx := context.WithValue(context.Background(), types.Trigger, types.CRON_JOB)

// 			go func() {
// 				release, err = statusService.GetRelease(ctx, "")
// 			}()

// 			time.Sleep(1 * time.Second)
// 			fmt.Println("2")

// 			It("should state is fetching", func() {
// 				fmt.Println("3")

// 				value, msg := service.GetStatus()
// 				Expect(value.Status).To(Equal(codegen.FetchUpdating))
// 				Expect(msg).To(Equal(""))
// 			})

// 			time.Sleep(2 * time.Second)

// 			It("should version is 0.4.8", func() {
// 				Expect(err).ToNot(HaveOccurred())
// 				Expect(release.Version).To(Equal("v0.4.8"))
// 			})

// 			It("should state is out-to-date", func() {
// 				value, msg := service.GetStatus()
// 				Expect(value.Status).To(Equal(codegen.Idle))
// 				Expect(msg).To(Equal("up-to-date"))
// 			})
// 		})
// 	})

// 	Describe("Install Update", func() {
// 		Context("Online Update", func() {
// 			It("should have github action", func() {

// 			})
// 		})
// 	})

// })
