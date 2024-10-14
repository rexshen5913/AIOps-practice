package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/internal/domain"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

type AIOpsRepository struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
}

func NewAIOpsRepository(clientset kubernetes.Interface, dynamicClient dynamic.Interface) *AIOpsRepository {
	return &AIOpsRepository{
		clientset:     clientset,
		dynamicClient: dynamicClient,
	}
}

func (r *AIOpsRepository) ListAIOpsResource(ctx context.Context, kind string) ([]domain.AIOps, error) {
	// 取得 API 群組和資源映射
	// RESTMapper
	discoveryClient := r.clientset.Discovery()
	// 從 API Server 獲取所有資源群組與版本
	apiGroupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		panic(err)
	}

	// 使用 discovery 資料建構 RESTMapper
	mapper := restmapper.NewDiscoveryRESTMapper(apiGroupResources)

	gvk := schema.GroupVersionKind{
		Group:   "aiops.geektime.com",
		Version: "v1alpha1",
		Kind:    kind,
	}

	// 將 GVK 轉換為 GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to map GVK to GVR: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	resourceInterface := r.dynamicClient.Resource(mapping.Resource).Namespace("default")
	resource, err := resourceInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	var aiopsList []domain.AIOps
	for _, item := range resource.Items {
		aiopsList = append(aiopsList, domain.AIOps{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			UID:       string(item.GetUID()),
		})
	}

	return aiopsList, nil

}
