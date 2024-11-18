export let structure = {
  "Cluster": [
    { group: "", version: "v1", resource: "namespaces", },
    { group: "", version: "v1", resource: "nodes" },
    { group: "", version: "v1", resource: "events" },
  ],
  "Workload": [
    { group: "", version: "v1", resource: "pods" },

    { group: "apps", version: "v1", resource: "deployments" },
    { group: "apps", version: "v1", resource: "statefulsets" },
    { group: "apps", version: "v1", resource: "daemonsets" },
    { group: "apps", version: "v1", resource: "replicasets" },

    { group: "batch", version: "v1", resource: "jobs" },
    { group: "batch", version: "v1", resource: "cronjobs" },
  ],
  "Config": [
    { group: "", version: "v1", resource: "configmaps" },
    { group: "", version: "v1", resource: "secrets" },
    { group: "autoscaling", version: "v2", resource: "horizontalpodautoscalers" },
    { group: "policy", version: "v1", resource: "poddisruptionbudgets" },
  ],
  "Network": [
    { group: "", version: "v1", resource: "services" },
    { group: "networking.k8s.io", version: "v1", resource: "networkpolicies" },
    { group: "networking.k8s.io", version: "v1", resource: "ingresses" },
    { group: "networking.k8s.io", version: "v1", resource: "ingressclasses" },
  ],
  "Storage": [
    { group: "storage.k8s.io", version: "v1", resource: "storageclasses" },
    { group: "", version: "v1", resource: "persistentvolumes" },
    { group: "", version: "v1", resource: "persistentvolumeclaims" },
  ],

  "Access Control": [
    { group: "rbac.authorization.k8s.io", version: "v1", resource: "roles" },
    { group: "rbac.authorization.k8s.io", version: "v1", resource: "rolebindings" },
    { group: "rbac.authorization.k8s.io", version: "v1", resource: "clusterroles" },
    { group: "rbac.authorization.k8s.io", version: "v1", resource: "clusterrolebindings" },
    { group: "", version: "v1", resource: "serviceaccounts" },
  ]
} as { [category: string]: { group: string, version: string, resource: string }[] };

export let hidden = [
  { group: "authentication.k8s.io", version: "v1", resource: "tokenreviews" },
  { group: "authentication.k8s.io", version: "v1", resource: "selfsubjectreviews" },
  { group: "authorization.k8s.io", version: "v1", resource: "subjectaccessreviews" },
  { group: "authorization.k8s.io", version: "v1", resource: "selfsubjectaccessreviews" },
  { group: "authorization.k8s.io", version: "v1", resource: "localsubjectaccessreviews" },
  { group: "authorization.k8s.io", version: "v1", resource: "selfsubjectrulesreviews" },
  { group: "", version: "v1", resource: "bindings" },
  { group: "", version: "v1", resource: "componentstatuses" },
  { group: "metrics.k8s.io", version: "v1beta1", resource: "pods" },
  { group: "metrics.k8s.io", version: "v1beta1", resource: "nodes" },
]
