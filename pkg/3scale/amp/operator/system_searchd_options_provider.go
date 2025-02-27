package operator

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
)

type SystemSearchdOptionsProvider struct {
	apimanager *appsv1alpha1.APIManager
	options    *component.SystemSearchdOptions
}

func NewSystemSearchdOptionsProvider(apimanager *appsv1alpha1.APIManager) *SystemSearchdOptionsProvider {
	return &SystemSearchdOptionsProvider{
		apimanager: apimanager,
		options:    component.NewSystemSearchdOptions(),
	}
}

func (s *SystemSearchdOptionsProvider) GetOptions() (*component.SystemSearchdOptions, error) {
	s.options.ImageTag = product.ThreescaleRelease
	s.options.Labels = s.labels()
	s.options.PodTemplateLabels = s.podTemplateLabels()
	s.setResourceRequirementsOptions()
	s.setNodeAffinityAndTolerationsOptions()
	s.setPVCOptions()

	err := s.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetSystemOptions validating: %w", err)
	}

	return s.options, nil
}

func (s *SystemSearchdOptionsProvider) setResourceRequirementsOptions() {
	s.options.ContainerResourceRequirements = v1.ResourceRequirements{}
	if *s.apimanager.Spec.ResourceRequirementsEnabled {
		s.options.ContainerResourceRequirements = component.DefaultSearchdContainerResourceRequirements()
	}
	// DeploymentConfig-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if s.apimanager.Spec.System.SearchdSpec.Resources != nil {
		s.options.ContainerResourceRequirements = *s.apimanager.Spec.System.SearchdSpec.Resources
	}
}

func (s *SystemSearchdOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	s.options.Affinity = s.apimanager.Spec.System.SearchdSpec.Affinity
	s.options.Tolerations = s.apimanager.Spec.System.SearchdSpec.Tolerations
}

func (s *SystemSearchdOptionsProvider) commonLabels() map[string]string {
	return map[string]string{
		"app":                  *s.apimanager.Spec.AppLabel,
		"threescale_component": "system",
	}
}

func (s *SystemSearchdOptionsProvider) labels() map[string]string {
	labels := s.commonLabels()
	labels["threescale_component_element"] = "searchd"
	return labels
}

func (s *SystemSearchdOptionsProvider) podTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("system-searchd", helper.ApplicationType)

	for k, v := range s.labels() {
		labels[k] = v
	}

	labels["deploymentConfig"] = "system-searchd"

	return labels
}

func (s *SystemSearchdOptionsProvider) setPVCOptions() {
	// Default values
	s.options.PVCOptions = component.SearchdPVCOptions{
		StorageClass:    nil,
		VolumeName:      "",
		StorageRequests: resource.MustParse("1Gi"),
	}

	if s.apimanager.Spec.System != nil &&
		s.apimanager.Spec.System.SearchdSpec != nil &&
		s.apimanager.Spec.System.SearchdSpec.PVC != nil {
		if s.apimanager.Spec.System.SearchdSpec.PVC.StorageClassName != nil {
			s.options.PVCOptions.StorageClass = s.apimanager.Spec.System.SearchdSpec.PVC.StorageClassName
		}
		if s.apimanager.Spec.System.SearchdSpec.PVC.Resources != nil {
			s.options.PVCOptions.StorageRequests = s.apimanager.Spec.System.SearchdSpec.PVC.Resources.Requests
		}
		if s.apimanager.Spec.System.SearchdSpec.PVC.VolumeName != nil {
			s.options.PVCOptions.VolumeName = *s.apimanager.Spec.System.SearchdSpec.PVC.VolumeName
		}
	}
}
