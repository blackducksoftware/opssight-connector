from cluster_clients import *

def create_hub_yaml():
    ns, err = subprocess.Popen(["./create-hub-yaml.sh"], stdout=subprocess.PIPE).communicate()
    return err

def create_opsight_yaml(arg1=""):
    ns, err = subprocess.Popen(["./create-opssight-yaml.sh",arg1], stdout=subprocess.PIPE).communicate()
    return err

def tests(k8s_client, hub_client, opssight_client):
    '''ns, err = subprocess.Popen(["./cpprog engsreepath471-engsreepath471.10.1.176.130.xip.io"], shell=True, stdout=subprocess.PIPE).communicate()
    if int(ns) > 0:
        print("Connected to Hub!")
    else:
        print("Couldn't connect")'''

    code_locs_json, err = subprocess.Popen("curl -sX GET -H 'Accept: application/json' http://perceptor-ops.10.1.176.68.xip.io/model", shell=True, stdout=subprocess.PIPE).communicate()
    d = json.loads(code_locs_json)
    print(d)
    

def main():
    # Create Project in the Cluster
    ns, err = subprocess.Popen(["oc", "new-project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    # Create yaml files
    create_hub_yaml()
    create_opsight_yaml("300m")
    # Push yaml files to the Cluster
    ns, err = subprocess.Popen(["oc", "create", "-f", "hub.yml"], stdout=subprocess.PIPE).communicate()
    if err != None:
        error(err)
    #ns, err = subprocess.Popen(["oc", "create", "-f", "opssight.yml"], stdout=subprocess.PIPE).communicate()
    print(err)
    # Create Clients to access Cluster Data
    k8s_client = K8sClient()
    hub_client = HubClient("aci-471-aci-471.10.1.176.130.xip.io")
    opssight_client = OpsSightClient()
    # Check if Scans are being performed
    tests(k8s_client, hub_client, opssight_client)
    # Tearing down the project
    ns, err = subprocess.Popen(["oc", "delete", "-f", "hub.yml"], stdout=subprocess.PIPE).communicate()
    #ns, err = subprocess.Popen(["oc", "delete", "-f", "opssight.yml"], stdout=subprocess.PIPE).communicate()
    ns, err = subprocess.Popen(["oc", "delete", "project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    return 0


main()