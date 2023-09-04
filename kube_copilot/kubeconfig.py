import base64
import os


def get_kubeconfig():
    token = open("/run/secrets/kubernetes.io/serviceaccount/token").read().strip()  # Strip newline characters
    cert = open("/run/secrets/kubernetes.io/serviceaccount/ca.crt").read().strip()  # Strip newline characters
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


def setup_kubeconfig():
    if not os.getenv("KUBERNETES_SERVICE_HOST"):
        # Not running inside a Pod, so no need to set up kubeconfig
        return

    home = os.path.expanduser("~")  # Use expanduser to get user's home directory
    kubeconfig_path = os.path.join(home, ".kube")
    kubeconfig_file = os.path.join(kubeconfig_path, "config")

    # If kubeconfig already exists, no need to recreate it
    if os.path.exists(kubeconfig_file):
        return

    os.makedirs(kubeconfig_path, exist_ok=True)
    kubeconfig = get_kubeconfig()
    with open(kubeconfig_file, "w") as f:
        f.write(kubeconfig)


# Call the setup_kubeconfig function to set up kubeconfig if needed
setup_kubeconfig()
