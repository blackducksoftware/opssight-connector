import json
import sys 
from cluster_clients import *
import logging

STATUS = b'UNKNOWN'


def diff_hub_IDs_and_opssight_IDs(hub_IDs, opssight_IDs):
    hub_IDs_set = set(hub_IDs)
    opssight_IDs_set = set(opssight_IDs)
    
    IDs_hub_only = hub_IDs_set.difference(opssight_IDs_set)
    IDs_both = hub_IDs_set.intersection(opssight_IDs_set)
    IDs_opssight_only = opssight_IDs_set.difference(hub_IDs_set)

    return IDs_hub_only, IDs_both, IDs_opssight_only


def assess_opssight(hubs, opssight, k8s):
    print("\n")
    # Get Total Images in Cluster
    cluster_images = k8s.get_images()
    print("Images In Your Kubernetes Cluster :", len(cluster_images))

    # Get Scanned Images from OpsSight
    opssight_IDs = opssight.get_shas_names()
    num_opssight_IDs = len(opssight_IDs)
    print("Images in OpsSight:", num_opssight_IDs)

    # Get Scanned Images from Hubs
    total_hub_IDs = []
    num_total_hub_IDs = 0
    for hub in hubs:
        hub_IDs = hub.get_code_locations_names()
        print("Total Images Scanned in Hub1 :", len(hub_IDs))
        num_total_hub_IDs += len(hub_IDs)
        total_hub_IDs.extend(hub_IDs)
    print("Cumulative Images Scanned by Hubs:", num_total_hub_IDs)

    # Analyze Scanned Images
    IDs_hub_only, IDs_hub_and_opssight, IDs_opssight_only = diff_hub_IDs_and_opssight_IDs(total_hub_IDs, opssight_IDs)

    # Display Ship-It Policy Results
    print("\n")
    print("***************************************")
    print("OpsSight Test Policy : "+str(100)+"% Image coverage to ship.")
    coverage = len(IDs_opssight_only) / float(len(cluster_images))
    print("OpsSight Coverage : %.2f%% Image Coverage" % coverage)
    print("***************************************")
    if coverage < 100.0:
        print("OpsSight Test Result: No Automated Release is Possible at this time.")
    else:
        print("OpsSight Test Result: PASS")
    print("***************************************")
    

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
    k8s_client = K8sClientWrapper()
    opssight_client = OpsSightClient(opssight_url)
    hub_client1 = HubClient(hub_url, usr, password)
    hub_client2 = HubClient("int-eric-int-eric.10.1.176.130.xip.io", usr, password)
    hub_client3 = HubClient("jim-emea-scaffold-jim-emea-scaffold.10.1.176.130.xip.io", usr, password)
    hub_client4 = HubClient("hammerp-hammerp.10.1.176.130.xip.io", usr, password)

    # TO DO: Testing...

    # Display OpsSight Assessment Test
    assess_opssight([hub_client1,hub_client2,hub_client3,hub_client4], opssight_client, k8s_client)


main()