/*
Copyright 2025.
*/

package v1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("HeliosApp Webhook Validation", func() {
	Context("Default() function", func() {
		It("should set default values for empty fields", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       0, // Should be defaulted to 1
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
					// gitBranch, gitopsBranch, gitopsPath should be defaulted
				},
			}

			heliosApp.Default()

			Expect(heliosApp.Spec.GitBranch).To(Equal("main"))
			Expect(heliosApp.Spec.GitopsBranch).To(Equal("main"))
			Expect(heliosApp.Spec.GitopsPath).To(Equal("test-app"))
			Expect(heliosApp.Spec.Replicas).To(Equal(int32(1)))
		})

		It("should not override existing values", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       3,
					GitBranch:      "develop",
					GitopsBranch:   "staging",
					GitopsPath:     "custom-path",
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			heliosApp.Default()

			Expect(heliosApp.Spec.GitBranch).To(Equal("develop"))
			Expect(heliosApp.Spec.GitopsBranch).To(Equal("staging"))
			Expect(heliosApp.Spec.GitopsPath).To(Equal("custom-path"))
			Expect(heliosApp.Spec.Replicas).To(Equal(int32(3)))
		})
	})

	Context("ValidateCreate() function", func() {
		It("should pass validation for valid HeliosApp", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			warnings, err := heliosApp.ValidateCreate()
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeEmpty())
		})

		It("should reject invalid Git URLs", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "invalid-url",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateCreate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("gitRepo is not a valid URL"))
		})

		It("should reject invalid image repository format", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx::invalid",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateCreate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("imageRepo has invalid format"))
		})

		It("should reject invalid port range", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           70000, // Invalid port
					Replicas:       1,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateCreate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("port must be between 1 and 65535"))
		})

		It("should reject negative replicas", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       -1, // Invalid replicas
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateCreate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("replicas cannot be negative"))
		})

		It("should reject empty required fields", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "", // Empty required field
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateCreate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("serviceAccount cannot be empty"))
		})

		It("should reject GitOps path starting with slash", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					GitopsPath:     "/invalid-path",
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateCreate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("gitopsPath should not start with"))
		})

		It("should reject GitOps path with ..", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					GitopsPath:     "app/../other",
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateCreate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("gitopsPath cannot contain"))
		})

		It("should reject branch names with whitespace", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					GitBranch:      "main branch", // Invalid with space
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateCreate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("gitBranch cannot contain whitespace"))
		})

		It("should warn for high replica count", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       15, // High replica count
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			warnings, err := heliosApp.ValidateCreate()
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(1))
			Expect(warnings[0]).To(ContainSubstring("high replica count"))
		})

		It("should handle multiple validation errors", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "invalid-url",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx::invalid",
					Port:           70000,
					Replicas:       -1,
					ServiceAccount: "",
					WebhookSecret:  "",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateCreate()
			Expect(err).To(HaveOccurred())
			errMsg := err.Error()
			Expect(errMsg).To(ContainSubstring("gitRepo is not a valid URL"))
			Expect(errMsg).To(ContainSubstring("imageRepo has invalid format"))
			Expect(errMsg).To(ContainSubstring("port must be between 1 and 65535"))
			Expect(errMsg).To(ContainSubstring("replicas cannot be negative"))
			Expect(errMsg).To(ContainSubstring("serviceAccount cannot be empty"))
			Expect(errMsg).To(ContainSubstring("webhookSecret cannot be empty"))
		})
	})

	Context("ValidateUpdate() function", func() {
		It("should use same validation logic as ValidateCreate", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "invalid-url",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			_, err := heliosApp.ValidateUpdate(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("gitRepo is not a valid URL"))
		})
	})

	Context("ValidateDelete() function", func() {
		It("should always pass", func() {
			heliosApp := &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/user/repo.git",
					GitopsRepo:     "https://github.com/user/gitops.git",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "pvc",
				},
			}

			warnings, err := heliosApp.ValidateDelete()
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeNil())
		})
	})

	Context("validateGitURL helper function", func() {
		It("should accept valid HTTPS URLs", func() {
			validURLs := []string{
				"https://github.com/user/repo.git",
				"https://gitlab.com/user/repo.git",
				"http://github.com/user/repo.git",
				"ssh://git@github.com:user/repo.git",
				"git://github.com/user/repo.git",
			}

			for _, url := range validURLs {
				heliosApp := &HeliosApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-app",
					},
					Spec: HeliosAppSpec{
						GitRepo:        url,
						GitopsRepo:     "https://github.com/user/gitops.git",
						ImageRepo:      "nginx:latest",
						Port:           8080,
						Replicas:       1,
						ServiceAccount: "pipeline-sa",
						WebhookSecret:  "webhook-secret",
						PVCName:        "pvc",
					},
				}

				_, err := heliosApp.ValidateCreate()
				Expect(err).NotTo(HaveOccurred(), "URL %s should be valid", url)
			}
		})

		It("should reject invalid URLs", func() {
			invalidURLs := []string{
				"invalid-url",
				"htp://github.com/user/repo.git", // Typo in http
				"ftp://github.com/user/repo.git", // Unsupported scheme
				"",                               // Empty URL
			}

			for _, url := range invalidURLs {
				heliosApp := &HeliosApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-app",
					},
					Spec: HeliosAppSpec{
						GitRepo:        url,
						GitopsRepo:     "https://github.com/user/gitops.git",
						ImageRepo:      "nginx:latest",
						Port:           8080,
						Replicas:       1,
						ServiceAccount: "pipeline-sa",
						WebhookSecret:  "webhook-secret",
						PVCName:        "pvc",
					},
				}

				_, err := heliosApp.ValidateCreate()
				Expect(err).To(HaveOccurred(), "URL %s should be invalid", url)
			}
		})
	})

	Context("validateImageRepo helper function", func() {
		It("should accept valid image repositories", func() {
			validImages := []string{
				"nginx",
				"nginx:latest",
				"docker.io/library/nginx",
				"gcr.io/project/image:v1.0.0",
				"registry.example.com/namespace/repo:tag",
			}

			for _, image := range validImages {
				heliosApp := &HeliosApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-app",
					},
					Spec: HeliosAppSpec{
						GitRepo:        "https://github.com/user/repo.git",
						GitopsRepo:     "https://github.com/user/gitops.git",
						ImageRepo:      image,
						Port:           8080,
						Replicas:       1,
						ServiceAccount: "pipeline-sa",
						WebhookSecret:  "webhook-secret",
						PVCName:        "pvc",
					},
				}

				_, err := heliosApp.ValidateCreate()
				Expect(err).NotTo(HaveOccurred(), "Image %s should be valid", image)
			}
		})

		It("should reject invalid image repositories", func() {
			invalidImages := []string{
				"",                // Empty
				"nginx::latest",   // Double colon
				"nginx:tag:extra", // Too many colons
				"nginx: ",         // Space in tag
				":latest",         // Empty repository
			}

			for _, image := range invalidImages {
				heliosApp := &HeliosApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-app",
					},
					Spec: HeliosAppSpec{
						GitRepo:        "https://github.com/user/repo.git",
						GitopsRepo:     "https://github.com/user/gitops.git",
						ImageRepo:      image,
						Port:           8080,
						Replicas:       1,
						ServiceAccount: "pipeline-sa",
						WebhookSecret:  "webhook-secret",
						PVCName:        "pvc",
					},
				}

				_, err := heliosApp.ValidateCreate()
				Expect(err).To(HaveOccurred(), "Image %s should be invalid", image)
			}
		})
	})
})
