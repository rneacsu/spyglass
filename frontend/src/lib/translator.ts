const languageFallbacks: { [lang: string]: string} = {
  "en-gb": "en",
  "en-us": "en",
}

const fallbackLanguage = "en";

const translations: { [lang: string]: { [key: string]: string } } = {
  "en": {
    "resource::/v1::namespaces": "Namespaces",
    "resource::/v1::nodes": "Nodes",
    "resource::/v1::events": "Events",
    "resource::/v1::pods": "Pods",
    "resource::apps/v1::deployments": "Deployments",
    "resource::apps/v1::statefulsets": "Stateful Sets",
    "resource::apps/v1::daemonsets": "Daemon Sets",
    "resource::apps/v1::replicasets": "Replica Sets",
    "resource::batch/v1::jobs": "Jobs",
    "resource::batch/v1::cronjobs": "Cron Jobs",
    "resource::/v1::configmaps": "Config Maps",
    "resource::/v1::secrets": "Secrets",
    "resource::autoscaling/v2::horizontalpodautoscalers": "Horizontal Pod Autoscalers",
    "resource::policy/v1::poddisruptionbudgets": "Pod Disruption Budgets",
    "resource::/v1::services": "Services",
    "resource::networking.k8s.io/v1::networkpolicies": "Network Policies",
    "resource::networking.k8s.io/v1::ingresses": "Ingresses",
    "resource::networking.k8s.io/v1::ingressclasses": "Ingress Classes",
    "resource::storage.k8s.io/v1::storageclasses": "Storage Classes",
    "resource::/v1::persistentvolumes": "Persistent Volumes",
    "resource::/v1::persistentvolumeclaims": "Persistent Volume Claims",
    "resource::rbac.authorization.k8s.io/v1::roles": "Roles",
    "resource::rbac.authorization.k8s.io/v1::rolebindings": "Role Bindings",
    "resource::rbac.authorization.k8s.io/v1::clusterroles": "Cluster Roles",
    "resource::rbac.authorization.k8s.io/v1::clusterrolebindings": "Cluster Role Bindings",
    "resource::/v1::serviceaccounts": "Service Accounts",
    "tableColumn::MinPods": "Min Pods",
    "tableColumn::MaxPods": "Max Pods",
    "tableCell::CrashLoopBackOff": "Crash Loop",
  }
}

export function translate(str: string, def: string | undefined = undefined): string {
  const language = navigator.language.toLowerCase();

  for (const lang of [language, languageFallbacks[language] ?? "n/a", fallbackLanguage]) {
    const translation = translations[lang]?.[str];
    if (translation) {
      return translation;
    }
  }

  return def ?? str;
}

export function translateTableColumn(str: string): string {
  return translate(`tableColumn::${str}`, str);
}

export function translateTableCell(str: string): string {
  return translate(`tableCell::${str}`, str);
}

export function translateResource(str: string): string {
  return translate(`resource::${str}`, str);
}
