def create_hub_yaml():
    ns, err = subprocess.Popen(["./create-hub.sh",arg1], stdout=subprocess.PIPE).communicate()
    return err

def create_opsight_yaml(arg1=""):
    ns, err = subprocess.Popen(["./create-opsight.sh",arg1], stdout=subprocess.PIPE).communicate()
    return err

def checks():
    pass

def main():
    # Create Project in the Cluster
    ns, err = subprocess.Popen(["oc", "new-project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    # Create yaml files
    create_hub_yaml()
    create_opsight_yaml("300m")
    # Push yaml files to the Cluster
    ns, err = subprocess.Popen(["oc", "create", "-f", "create-hub.yml"], stdout=subprocess.PIPE).communicate()
    ns, err = subprocess.Popen(["oc", "create", "-f", "create-opssight.yml"], stdout=subprocess.PIPE).communicate()
    # Check if Scans are being performed
    checks()
    # Tearing down the project
    ns, err = subprocess.Popen(["oc", "delete", "-f", "create-hub.yml"], stdout=subprocess.PIPE).communicate()
    ns, err = subprocess.Popen(["oc", "delete", "-f", "create-opssight.yml"], stdout=subprocess.PIPE).communicate()
    ns, err = subprocess.Popen(["oc", "delete", "project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    return 0