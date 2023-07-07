import base64
import os


def get_kubeconfig():
    token = open("/run/secrets/kubernetes.io/serviceaccount/token").read()
    cert = open("/run/secrets/kubernetes.io/serviceaccount/ca.crt").read()
    cert = base64.b64encode(cert.encode()).decode()
    host = os.environ.get("KUBERNETES_SERVICE_HOST")
    port = os.environ.get("KUBERNETES_SERVICE_PORT")

    return f'''apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {cert}
    server: https://{host}:{port}
  name: kube
contexts:
- context:
    cluster: kube
    user: kube
  name: kube
current-context: kube
kind: Config
users:
- name: kube
  user:
    token: {token}
'''


def write_kubeconfig(kubeconfig):
    if not os.getenv("KUBERNETES_SERVICE_HOST", None):
        # not running inside Pod
        return

    home = os.environ.get("HOME")
    kubeconfig_path = os.path.join(home, ".kube")
    kubeconfig_file = os.path.join(kubeconfig_path, "config")

    # kubeconfig already exists
    if os.path.exists(kubeconfig_file):
        return

    os.makedirs(kubeconfig_path, exist_ok=True)
    with open(kubeconfig_file, "w") as f:
        f.write(kubeconfig)


def setup_kubeconfig():
    write_kubeconfig(get_kubeconfig())
