apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: cpu-limit.demo.webhook
webhooks:
- name: cpu-limit.demo.webhook
  namespaceSelector:
    matchExpressions:
    - key: mutatingwebhook
      operator: In
      values:
      - active
  rules:
  - apiGroups:
    - "apps"
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - deployments
  failurePolicy: Fail
  clientConfig:
    service:
      namespace: kube-system
      name: admission-webhook
    caBundle: ${CABUNDLE}
