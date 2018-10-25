from cluster_clients import *

def tests(k8s_client, hub_client, opssight_client):
    if len(hub_client.get_projects_names()) > 0:
        print("Connected to Hub!")
    else:
        print("Couldn't Connect to Hub")
        '''print("Tearing down hub, opsSight, and project")
        #hub_client.destory()
        opssight_client.destroy()
        ns, err = subprocess.Popen(["oc", "delete", "project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()'''
        sys.exit(1)

    if len(opssight_client.get_shas_names()) > 0:
        print("Connected to OpsSight!")
    else:
        print("Couldn't Connect to OpsSight")
        '''print("Tearing down hub, opsSight, and project")
        #hub_client.destory()
        opssight_client.destroy()
        ns, err = subprocess.Popen(["oc", "delete", "project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()'''
        sys.exit(1)
    

def main():
    # path to config
    test_config_path = sys.argv[1]
    test_config_json = None
    with open(test_config_path) as f:
        test_config_json = json.load(f)
    opssight_url = test_config_json.opssight_url 
    hub_url = test_config_json.hub_url 

    # Create Project in the Cluster
    '''ns, err = subprocess.Popen(["oc", "new-project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    if err != None:
        sys.exit(err)'''

    # Create Clients to access Cluster Data
    k8s_client = K8sClientWrapper()
    #hub_client = HubClient('engsreepath471-engsreepath471.10.1.176.130.xip.io')
    opssight_client = OpsSightClient(k8s_client)

    # Put a Hub and OpsSight into the Cluster
    #hub_client.create()
    '''opssight_client.create()'''

    # Check if Scans are being performed
    tests(k8s_client, hub_client, opssight_client)

    # Tearing down the project
    '''print("Tearing down the Hub")
    #hub_client.destory()
    print("Tearing down OpsSight")
    opssight_client.destroy()
    print("Tearing Project")
    ns, err = subprocess.Popen(["oc", "delete", "project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()'''

    return 0


main()