package controller

import (
	"testing"

	homerv1alpha1 "github.com/jplanckeel/homer-k8s/api/v1alpha1"
)

func TestHomerServicesController(t *testing.T) {
	config := homerv1alpha1.HomerServicesList{
		Items: []homerv1alpha1.HomerServices{
			{
				Spec: homerv1alpha1.HomerServicesSpec{
					Page: "",
					Groups: []homerv1alpha1.Group{
						{
							Name: "Home",
							Icon: "fas fa-home",
							Items: []homerv1alpha1.Item{
								{
									Name: "Home",
									Icon: "fas fa-home",
									Url:  "https://www.service1.example.com",
								},
							},
						},
					},
				},
			},
			{
				Spec: homerv1alpha1.HomerServicesSpec{
					Page: "Page 1",
					Groups: []homerv1alpha1.Group{
						{
							Name: "Home",
							Icon: "fas fa-home",
							Items: []homerv1alpha1.Item{
								{
									Name: "Home",
									Icon: "fas fa-home",
									Url:  "https://www.service2.example.com",
								},
							},
						},
					},
				},
			},
			{
				Spec: homerv1alpha1.HomerServicesSpec{
					Page: "Page 1",
					Groups: []homerv1alpha1.Group{
						{
							Name: "Test",
							Icon: "fas fa-home",
							Items: []homerv1alpha1.Item{
								{
									Name: "Example",
									Icon: "fas fa-home",
									Url:  "https://www.service3.example.com",
								},
							},
						},
					},
				},
			},
			{
				Spec: homerv1alpha1.HomerServicesSpec{
					Page: "Page 2",
					Groups: []homerv1alpha1.Group{
						{
							Name: "Home",
							Icon: "fas fa-home",
							Items: []homerv1alpha1.Item{
								{
									Name: "Home",
									Icon: "fas fa-home",
									Url:  "https://www.service4.example.com",
								},
							},
						},
					},
				},
			},
		},
	}

	t.Run("TestSplitServicesPerPage", func(t *testing.T) {
		pages, groups := splitServicesPerPage(&config)

		t.Logf("Pages: %v", pages)
		t.Logf("Groups: %v", groups)

		if len(pages) != 3 {
			t.Errorf("Expected 3 pages, but got %d", len(pages))
		}
		if len(groups) != 3 {
			t.Errorf("Expected 3 groups, but got %d", len(groups))
		}
	})

	// Add more test cases for other fields...

}
