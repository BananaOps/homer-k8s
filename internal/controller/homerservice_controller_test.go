package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	homerv1alpha1 "github.com/jplanckeel/homer-k8s/api/v1alpha1"
)

func TestMergeGroupWithSameName(t *testing.T) {
	groups := []homerv1alpha1.Group{
		{
			Name: "group1",
			Icon: "",
			Items: []homerv1alpha1.Item{
				{
					Name: "item1",
				},
			},
		},
		{
			Name: "group2",
			Items: []homerv1alpha1.Item{
				{
					Name: "item2",
				},
			},
		},
		{
			Name: "group1",
			Items: []homerv1alpha1.Item{
				{
					Name: "item3",
				},
			},
		},
	}

	expected := []homerv1alpha1.Group{
		{
			Name: "group1",
			Items: []homerv1alpha1.Item{
				{
					Name: "item1",
				},
				{
					Name: "item3",
				},
			},
		},
		{
			Name: "group2",
			Items: []homerv1alpha1.Item{
				{
					Name: "item2",
				},
			},
		},
	}

	result := mergeGroupWithSameName(groups)

	assert.Equal(t, result, expected, "Test Merge Group")

}
