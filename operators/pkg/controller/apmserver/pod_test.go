// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package apmserver

import (
	"reflect"
	"testing"

	"github.com/elastic/cloud-on-k8s/operators/pkg/apis/apm/v1alpha1"
	"github.com/elastic/cloud-on-k8s/operators/pkg/controller/common/volume"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewPodSpec(t *testing.T) {
	configSecretVol := volume.NewSecretVolumeWithMountPath(
		"config-secret",
		"config",
		"/usr/share/apm-server/config/config-secret",
	)
	varFalse := false
	probe := readinessProbe(true)
	tests := []struct {
		name string
		as   v1alpha1.ApmServer
		p    PodSpecParams
		want corev1.PodTemplateSpec
	}{
		{
			name: "create default pod spec",
			as: v1alpha1.ApmServer{
				TypeMeta: metav1.TypeMeta{
					Kind: "ApmServer",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-apm",
					Namespace: "default",
				},
			},
			p: PodSpecParams{
				Version: "7.0.1",
				ConfigSecret: corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns",
						Name:      "config-secret",
					},
				},
				ApmServerSecret: corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "apm-secret",
					},
				},
			},
			want: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						configSecretVol.Volume(), configVolume.Volume(),
					},
					AutomountServiceAccountToken: &varFalse,
					Containers: []corev1.Container{
						{
							Name:  v1alpha1.APMServerContainerName,
							Image: imageWithVersion(defaultImageRepositoryAndName, "7.0.1"),
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"},
									},
								},
								{
									Name: "SECRET_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "apm-secret"},
											Key:                  SecretTokenKey,
										},
									},
								},
							},
							ReadinessProbe: &probe,
							Ports:          ports,
							Command:        command,
							VolumeMounts: []corev1.VolumeMount{
								configSecretVol.VolumeMount(), configVolume.VolumeMount(),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newPodSpec(&tt.as, tt.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPodSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}
