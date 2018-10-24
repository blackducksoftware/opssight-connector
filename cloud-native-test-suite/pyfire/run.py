from cluster_clients import *

def tests(k8s_client, hub_client, opssight_client):
    if len(hub_client.get_projects_names()) > 0:
        print("Connected to Hub!")
    else:
        sys.exit("Couldn't Connect to Hub")

    if len(opssight_client.get_shas_names()) > 0:
        print("Connected to OpsSight!")
    else:
        sys.exit("Couldn't Connect to OpsSight")
    

def main():
    # Create Project in the Cluster
    ns, err = subprocess.Popen(["oc", "new-project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    if err != None:
        sys.exit(err)

    # Create Clients to access Cluster Data
    k8s_client = K8sClient()
    hub_client = HubClient('engsreepath471-engsreepath471.10.1.176.130.xip.io')
    opssight_client = OpsSightClient()

    # Put a Hub and OpsSight into the Cluster
    #hub_client.create()
    opssight_client.create()

    # Check if Scans are being performed
    tests(k8s_client, hub_client, opssight_client)

    # Tearing down the project
    #hub_client.destory()
    opssight_client.destroy()
    ns, err = subprocess.Popen(["oc", "delete", "project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()

    return 0


main()