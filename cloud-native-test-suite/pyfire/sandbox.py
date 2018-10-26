import sys 
import os
import json
import subprocess
from cluster_clients import *

def main():
    if len(sys.argv) < 2:
        print("USAGE:")
        print("python3 run.py <config_file_path>")
        sys.exit("Wrong Number of Parameters")

    logging.basicConfig(level=logging.ERROR, stream=sys.stdout)
    logging.debug("Starting Tests")
    
    # Read parameters from config file
    test_config_path = sys.argv[1]

    test_config_json = None
    with open(test_config_path) as f:
        test_config_json = json.load(f)

    opssight_url = test_config_json["PerceptorURL"]
    hub_url = test_config_json["HubURL"]
    port = test_config_json["Port"]
    usr = test_config_json["Username"]
    password = test_config_json["Password"]

    # Create Kubernetes, OpsSight, and Hub Clients
    k8s_client = K8sClient()
    opssight_client = OpsSightClient(opssight_url, k8s_client)
    hub_client1 = HubClient(hub_url, k8s_client, usr, password)
    hub_client2 = HubClient("int-eric-int-eric.10.1.176.130.xip.io", k8s_client, usr, password)
    hub_client3 = HubClient("jim-emea-scaffold-jim-emea-scaffold.10.1.176.130.xip.io", k8s_client, usr, password)
    hub_client4 = HubClient("hammerp-hammerp.10.1.176.130.xip.io", k8s_client, usr, password)

    # TO DO - Add Functinality to the Hub
    print(hub_client1.get_projects_dump())


main()