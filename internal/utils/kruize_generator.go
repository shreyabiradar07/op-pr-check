package utils

import (
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
)

func boolPtr(b bool) *bool {
	return &b
}


// KruizeResourceGenerator holds common data needed for creating resources.
type KruizeResourceGenerator struct {
	Namespace string
	Autotune_image string
	Autotune_ui_image string
}

// NewKruizeResourceGenerator creates a new generator for Kruize resources.
func NewKruizeResourceGenerator(namespace string, autotuneImage string, autotuneUIImage string) *KruizeResourceGenerator {
	// If no image is provided from the CR, use a sensible default.
	if autotuneImage == "" {
		autotuneImage = "quay.io/kruize/autotune_operator:latest"
	}
    if autotuneUIImage == "" {
        autotuneUIImage = "quay.io/kruize/kruize-ui:0.0.9"
    }
	return &KruizeResourceGenerator{
		Namespace:   namespace,
		Autotune_image: autotuneImage,
		Autotune_ui_image : autotuneUIImage,
	}
}

// ClusterScopedResources generates all cluster-scoped resources for Kruize.
// These resources DO NOT get an owner reference.
func (g *KruizeResourceGenerator) ClusterScopedResources() []client.Object {
	return []client.Object{
		g.recommendationUpdaterClusterRole(),
		g.recommendationUpdaterClusterRoleBinding(),
		g.monitoringViewClusterRoleBinding(),
		// g.systemReaderClusterRoleBinding(),
		// g.monitoringAccessClusterRole(),
		// g.monitoringAccessClusterRoleBinding(),
		g.kruizeDBPersistentVolume(),
		g.kruizeDBPersistentVolumeClaim(),
	}
}

// NamespacedResources generates all namespaced resources for Kruize.
// These resources will get an owner reference set to the Kruize CR.
func (g *KruizeResourceGenerator) NamespacedResources() []client.Object {
	var objects []client.Object
	objects = append(objects, g.DBResources()...)
	objects = append(objects, g.BackendResources()...)
	objects = append(objects, g.UIResources()...)
	objects = append(objects, g.Routes()...)
	return objects
}

func (g *KruizeResourceGenerator) Routes() []client.Object {
	routes := []*routev1.Route{
		g.generateRoute("kruize", "kruize-service", "http"),
		g.generateRoute("kruize-ui-nginx-service", "kruize-ui-nginx-service", "http"),
	}

	objects := make([]client.Object, len(routes))
	for i, route := range routes {
		objects[i] = route
	}
	return objects
}

// generateRoute is a private helper to avoid duplicating Route creation logic.
func (g *KruizeResourceGenerator) generateRoute(name, serviceName, targetPort string) *routev1.Route {
	weight := int32(100)
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "route.openshift.io/v1",
			Kind:       "Route",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: g.Namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   serviceName,
				Weight: &weight,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString(targetPort),
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}
}

func (g *KruizeResourceGenerator) RBACAndConfigResources() []client.Object {
	return []client.Object{
		g.KruizeServiceAccount(),
		g.recommendationUpdaterClusterRole(),
		g.recommendationUpdaterClusterRoleBinding(),
		g.monitoringViewClusterRoleBinding(),
		g.prometheusReaderClusterRoleBinding(),
		g.systemReaderClusterRoleBinding(),
		g.monitoringAccessClusterRole(),
		g.monitoringAccessClusterRoleBinding(),
		g.KruizeConfigMap(),
	}
}


// kruizeServiceAccount generates the ServiceAccount for Kruize.
func (g *KruizeResourceGenerator) KruizeNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: g.Namespace,
		},
	}
}

// kruizeServiceAccount generates the ServiceAccount for Kruize.
func (g *KruizeResourceGenerator) KruizeServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kruize-sa",
			Namespace: g.Namespace,
		},
	}
}

// recommendationUpdaterClusterRole generates the ClusterRole for Kruize recommendation updater.
func (g *KruizeResourceGenerator) recommendationUpdaterClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kruize-recommendation-updater",
		},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{""}, Resources: []string{"pods", "nodes", "namespaces", "services", "endpoints"}, Verbs: []string{"get", "list", "watch"}},
			{APIGroups: []string{"apps"}, Resources: []string{"deployments", "replicasets", "statefulsets", "daemonsets"}, Verbs: []string{"get", "list", "watch"}},
			{APIGroups: []string{"extensions", "networking.k8s.io"}, Resources: []string{"ingresses"}, Verbs: []string{"get", "list", "watch"}},
			{APIGroups: []string{"autoscaling"}, Resources: []string{"horizontalpodautoscalers"}, Verbs: []string{"get", "list", "watch"}},
			{APIGroups: []string{"autoscaling.k8s.io"}, Resources: []string{"verticalpodautoscalers"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch"}},
			{APIGroups: []string{"metrics.k8s.io"}, Resources: []string{"pods", "nodes"}, Verbs: []string{"get", "list"}},
			{APIGroups: []string{"monitoring.coreos.com"}, Resources: []string{"prometheuses", "alertmanagers", "servicemonitors"}, Verbs: []string{"get", "list", "watch"}, ResourceNames: []string{"*"}},
			{APIGroups: []string{"monitoring.coreos.com"}, Resources: []string{"prometheuses/api"}, Verbs: []string{"get", "create", "update"}},
			{NonResourceURLs: []string{"/metrics", "/api/v1/label/*", "/api/v1/query*", "/api/v1/series*", "/api/v1/targets*"}, Verbs: []string{"get"}},
		},
	}
}

// recommendationUpdaterClusterRoleBinding generates the ClusterRoleBinding for the recommendation updater.
func (g *KruizeResourceGenerator) recommendationUpdaterClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kruize-recommendation-updater-crb",
		},
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: "kruize-sa", Namespace: g.Namespace},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "kruize-recommendation-updater",
		},
	}
}

// monitoringViewClusterRoleBinding generates the ClusterRoleBinding for cluster monitoring view.
func (g *KruizeResourceGenerator) monitoringViewClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kruize-monitoring-view",
		},
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: "kruize-sa", Namespace: g.Namespace},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-monitoring-view",
		},
	}
}

// prometheusReaderClusterRoleBinding generates the ClusterRoleBinding for prometheus reader.
func (g *KruizeResourceGenerator) prometheusReaderClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kruize-prometheus-reader",
		},
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: "kruize-sa", Namespace: g.Namespace},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "view",
		},
	}
}

// systemReaderClusterRoleBinding generates the ClusterRoleBinding for system reader.
func (g *KruizeResourceGenerator) systemReaderClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kruize-system-reader",
		},
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: "kruize-sa", Namespace: g.Namespace},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:monitoring",
		},
	}
}

// monitoringAccessClusterRole generates the ClusterRole for monitoring access.
func (g *KruizeResourceGenerator) monitoringAccessClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kruize-monitoring-access",
		},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{"monitoring.coreos.com"}, Resources: []string{"prometheuses", "prometheuses/api", "alertmanagers", "servicemonitors", "prometheusrules"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch"}},
			{NonResourceURLs: []string{"/api/v1/*", "/metrics"}, Verbs: []string{"get"}},
		},
	}
}

// monitoringAccessClusterRoleBinding generates the ClusterRoleBinding for monitoring access.
func (g *KruizeResourceGenerator) monitoringAccessClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kruize-monitoring-access-crb",
		},
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: "kruize-sa", Namespace: g.Namespace},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "kruize-monitoring-access",
		},
	}
}

// kruizeDBPersistentVolume generates the PersistentVolume for the Kruize database.
// Note: PersistentVolumes are cluster-scoped resources.
func (g *KruizeResourceGenerator) kruizeDBPersistentVolume() *corev1.PersistentVolume {
	return &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolume",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kruize-db-pv-volume",
			Labels: map[string]string{
				"type": "local",
				"app":  "kruize-db",
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: "manual",
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("500Mi"),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			// The HostPath must be nested inside the PersistentVolumeSource struct.
            PersistentVolumeSource: corev1.PersistentVolumeSource{
                HostPath: &corev1.HostPathVolumeSource{
                    Path: "/mnt/data",
                },
            },
		},
	}
}

// kruizeDBPersistentVolumeClaim generates the PersistentVolumeClaim for the Kruize database.
func (g *KruizeResourceGenerator) kruizeDBPersistentVolumeClaim() *corev1.PersistentVolumeClaim {
	storageClassName := "manual"
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kruize-db-pv-claim",
			Namespace: g.Namespace,
			Labels: map[string]string{
				"app": "kruize-db",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("500Mi"),
				},
			},
		},
	}
}

// kruizeConfigMap generates the main ConfigMap for Kruize.
func (g *KruizeResourceGenerator) KruizeConfigMap() *corev1.ConfigMap {
	dbConfigJSON := `{
      "database": {
        "adminPassword": "kruize123",
        "adminUsername": "kruize",
        "hostname": "kruize-db-service",
        "name": "kruizedb",
        "password": "kruize123",
        "port": 5432,
        "sslMode": "disable",
        "username": "kruize"
      }
    }`

	kruizeConfigJSON := fmt.Sprintf(`{
      "clustertype": "kubernetes",
      "k8stype": "openshift",
      "authtype": "",
      "monitoringagent": "prometheus",
      "monitoringservice": "prometheus-k8s",
      "monitoringendpoint": "prometheus-k8s",
      "savetodb": "true",
      "dbdriver": "jdbc:postgresql://",
      "plots": "true",
      "local": "true",
      "logAllHttpReqAndResp": "true",
      "recommendationsURL" : "http://kruize-service.%s.svc.cluster.local:8080/generateRecommendations?experiment_name=%%s",
      "experimentsURL" : "http://kruize-service.%s.svc.cluster.local:8080/createExperiment",
      "experimentNameFormat" : "%%datasource%%|%%clustername%%|%%namespace%%|%%workloadname%%(%%workloadtype%%)|%%containername%%",
      "bulkapilimit" : 1000,
      "isKafkaEnabled" : "false",
      "hibernate": {
        "dialect": "org.hibernate.dialect.PostgreSQLDialect",
        "driver": "org.postgresql.Driver",
        "c3p0minsize": 5,
        "c3p0maxsize": 10,
        "c3p0timeout": 300,
        "c3p0maxstatements": 100,
        "hbm2ddlauto": "none",
        "showsql": "false",
        "timezone": "UTC"
      },
      "datasource": [
        {
          "name": "prometheus-1",
          "provider": "prometheus",
          "namespace": "openshift-monitoring",
          "url": "https://prometheus-k8s.openshift-monitoring.svc.cluster.local:9091",
          "authentication": {
              "type": "bearer",
              "credentials": {
                "tokenFilePath": "/var/run/secrets/kubernetes.io/serviceaccount/token"
              }
          }
        }
      ]
    }`, g.Namespace, g.Namespace)

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kruizeconfig",
			Namespace: g.Namespace,
		},
		Data: map[string]string{
			"dbconfigjson":     dbConfigJSON,
			"kruizeconfigjson": kruizeConfigJSON,
		},
	}
}



// DBResources generates all resources related to the Kruize Database (PostgreSQL).
func (g *KruizeResourceGenerator) DBResources() []client.Object {
	return []client.Object{
		g.kruizeDBDeployment(),
		g.kruizeDBService(),
	}
}

// kruizeDBDeployment is a private helper that generates the Deployment for the database.
func (g *KruizeResourceGenerator) kruizeDBDeployment() *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kruize-db",
			Namespace: g.Namespace,
			Labels: map[string]string{
				"app": "kruize-db",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "kruize-db",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "kruize-db",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "kruize-sa", // Assuming this SA exists or is also generated
					Containers: []corev1.Container{
						{
							Name:            "kruize-db",
							Image:           "quay.io/kruizehub/postgres:15.2",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "POSTGRES_PASSWORD", Value: "kruize123"},
								{Name: "POSTGRES_USER", Value: "kruize"},
								{Name: "POSTGRES_DB", Value: "kruizedb"},
								{Name: "POSTGRES_INITDB_ARGS", Value: "--auth-host=md5"},
								{Name: "PGDATA", Value: "/tmp/pgdata"},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("200Mi"),
									corev1.ResourceCPU:    resource.MustParse("100m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
									corev1.ResourceCPU:    resource.MustParse("500m"),
								},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 5432},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "postgres-storage", MountPath: "/tmp"},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-c", "pg_isready -U kruize -d kruizedb"},
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-c", "pg_isready -U kruize -d kruizedb"},
									},
								},
								InitialDelaySeconds: 45,
								PeriodSeconds:       20,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "postgres-storage",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

// kruizeDBService is a private helper that generates the Service for the database.
func (g *KruizeResourceGenerator) kruizeDBService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kruize-db-service",
			Namespace: g.Namespace,
			Labels: map[string]string{
				"app": "kruize-db",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "kruize-db-port",
					Port:       5432,
					TargetPort: intstr.FromInt(5432),
				},
			},
			Selector: map[string]string{
				"app": "kruize-db",
			},
		},
	}
}


// BackendResources generates the core Kruize application Deployment and Service.
func (g *KruizeResourceGenerator) BackendResources() []client.Object {
	return []client.Object{
		g.kruizeDeployment(),
		g.kruizeService(),
	}
}

// kruizeDeployment is a private helper that generates the deployment for the Kruize backend.
func (g *KruizeResourceGenerator) kruizeDeployment() *appsv1.Deployment {
	replicas := int32(1)
	automountToken := true

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kruize",
			Namespace: g.Namespace,
			Labels: map[string]string{
				"app": "kruize",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "kruize",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "kruize",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:           "kruize-sa",
					AutomountServiceAccountToken: &automountToken,
					Containers: []corev1.Container{
						{
							Name:  "kruize",
							Image: g.Autotune_image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080},
							},
							Env: []corev1.EnvVar{
								{Name: "LOGGING_LEVEL", Value: "info"},
								{Name: "ROOT_LOGGING_LEVEL", Value: "error"},
								{Name: "DB_CONFIG_FILE", Value: "/etc/config/dbconfigjson"},
								{Name: "KRUIZE_CONFIG_FILE", Value: "/etc/config/kruizeconfigjson"},
								{Name: "JAVA_TOOL_OPTIONS", Value: "-XX:MaxRAMPercentage=80"},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("256Mi"),
									corev1.ResourceCPU:    resource.MustParse("250m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
									corev1.ResourceCPU:    resource.MustParse("500m"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "config-volume", MountPath: "/etc/config"},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 45,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 90,
								PeriodSeconds:       20,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "kruizeconfig",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// kruizeService is a private helper that generates the Service for the Kruize backend.
func (g *KruizeResourceGenerator) kruizeService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kruize-service",
			Namespace: g.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": "kruize",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}
}

// UIResources generates all resources related to the Kruize UI.
func (g *KruizeResourceGenerator) UIResources() []client.Object {
	// We will create helpers for the ConfigMap and Service next.
	return []client.Object{
	    g.nginxConfigMap(),
	    g.kruizeUINginxService(),
		g.kruizeUINginxPod(),
	}
}


func (g *KruizeResourceGenerator) kruizeUINginxPod() *corev1.Pod {
	return &corev1.Pod{
		// The TypeMeta tells the client which kind of object this is.
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		// The ObjectMeta contains the name, namespace, and labels.
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kruize-ui-nginx-pod",
			Namespace: g.Namespace, // We use the namespace from our generator struct.
			Labels: map[string]string{
				"app": "kruize-ui-nginx",
			},
		},
		// The Spec defines the desired state of the Pod.
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "kruize-ui-nginx-container",
					Image:           g.Autotune_ui_image,
					ImagePullPolicy: corev1.PullAlways,
					Env: []corev1.EnvVar{
						{Name: "KRUIZE_API_URL", Value: "http://kruize-service:8080"},
						{Name: "REACT_APP_KRUIZE_API_URL", Value: "http://kruize-service:8080"},
						{Name: "KRUIZE_UI_API_URL", Value: "http://kruize-service:8080"},
						{Name: "API_URL", Value: "http://kruize-service:8080"},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "nginx-config-volume",
							MountPath: "/etc/nginx/nginx.conf",
							SubPath:   "nginx.conf",
						},
                        {
                            Name: "nginx-cache",
                            MountPath: "/var/cache/nginx",
                        },
                        {Name: "nginx-pid", MountPath: "/var/run"},
                        {Name: "nginx-tmp", MountPath: "/tmp"},
					},
                    SecurityContext: &corev1.SecurityContext{
                        AllowPrivilegeEscalation: boolPtr(false),
                        RunAsNonRoot: boolPtr(true),
                        Capabilities: &corev1.Capabilities{
                            Drop: []corev1.Capability{"ALL"},
                        },
                        SeccompProfile: &corev1.SeccompProfile{
                            Type: corev1.SeccompProfileTypeRuntimeDefault,
                        },
                    },
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "nginx-config-volume",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "nginx-config",
							},
						},
					},
				},
                // Define the emptyDir volume that will be used for caching.
                {
                    Name: "nginx-cache",
                    VolumeSource: corev1.VolumeSource{
                        EmptyDir: &corev1.EmptyDirVolumeSource{},
                    },
                },
                {Name: "nginx-pid", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
                {Name: "nginx-tmp", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
		},
	}
}

// nginxConfigMap is an internal helper that generates the ConfigMap for the Nginx configuration.
func (g *KruizeResourceGenerator) nginxConfigMap() *corev1.ConfigMap {
	nginxConf := `
events {}
http {
  upstream kruize-api {
	server kruize-service:8080;
  }

  server {
	listen 8080;
	server_name localhost;

	root   /usr/share/nginx/html;

	location ^~ /api/ {
	  rewrite ^/api(.*)$ $1 break;
	  proxy_pass http://kruize-api;
	}

	location / {
	  index index.html;
	  error_page 404 =200 /index.html;
	}
  }
}
`
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-config",
			Namespace: g.Namespace,
		},
		Data: map[string]string{
			"nginx.conf": nginxConf,
		},
	}
}

// kruizeUINginxService is an internal helper that generates the Service for the Kruize UI Nginx pod.
func (g *KruizeResourceGenerator) kruizeUINginxService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kruize-ui-nginx-service",
			Namespace: g.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{
				"app": "kruize-ui-nginx",
			},
		},
	}
}