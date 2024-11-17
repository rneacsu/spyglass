const languageFallbacks: { [lang: string]: string} = {
  "en-gb": "en",
  "en-us": "en",
}

const fallbackLanguage = "en";

const translations: { [lang: string]: { [key: string]: string } } = {
  "en": {
    "/v1::namespaces": "Namespaces",
    "/v1::nodes": "Nodes",
    "/v1::events": "Events",
    "/v1::pods": "Pods",
    "apps/v1::deployments": "Deployments",
    "apps/v1::statefulsets": "Stateful Sets",
    "apps/v1::daemonsets": "Daemon Sets",
    "apps/v1::replicasets": "Replica Sets",
    "batch/v1::jobs": "Jobs",
    "batch/v1::cronjobs": "Cron Jobs",
    "/v1::configmaps": "Config Maps",
    "/v1::secrets": "Secrets",
    "autoscaling/v2::horizontalpodautoscalers": "Horizontal Pod Autoscalers",
    "policy/v1::poddisruptionbudgets": "Pod Disruption Budgets",
    "/v1::services": "Services",
    "networking.k8s.io/v1::networkpolicies": "Network Policies",
    "networking.k8s.io/v1::ingresses": "Ingresses",
    "networking.k8s.io/v1::ingressclasses": "Ingress Classes",
    "storage.k8s.io/v1::storageclasses": "Storage Classes",
    "/v1::persistentvolumes": "Persistent Volumes",
    "/v1::persistentvolumeclaims": "Persistent Volume Claims",
    "rbac.authorization.k8s.io/v1::roles": "Roles",
    "rbac.authorization.k8s.io/v1::rolebindings": "Role Bindings",
    "rbac.authorization.k8s.io/v1::clusterroles": "Cluster Roles",
    "rbac.authorization.k8s.io/v1::clusterrolebindings": "Cluster Role Bindings",
    "/v1::serviceaccounts": "Service Accounts",
  }
}

export function translate(str: string): string {
  const language = navigator.language.toLowerCase();

  for (const lang of [language, languageFallbacks[language] ?? "n/a", fallbackLanguage]) {
    const translation = translations[lang]?.[str];
    if (translation) {
      return translation;
    }
  }

  return str;
}
