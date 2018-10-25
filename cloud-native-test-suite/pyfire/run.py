import json
import sys 
from cluster_clients import *
from http.server import BaseHTTPRequestHandler, HTTPServer
import logging

STATUS = b'UNKNOWN'

class myHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type','text/html')
        self.end_headers()
        self.wfile.write(STATUS)

def diff_hub_IDs_and_opssight_IDs(hub_IDs, opssight_IDs):
    hub_IDs_set = set(hub_IDs)
    opssight_IDs_set = set(opssight_IDs)
    
    IDs_hub_only = hub_IDs_set.difference(opssight_IDs_set)
    IDs_both = hub_IDs_set.intersection(opssight_IDs_set)
    IDs_opssight_only = opssight_IDs_set.difference(hub_IDs_set)

    return IDs_hub_only, IDs_both, IDs_opssight_only
    
def compare_huub_IDs_and_opssight_IDs(raw_hub, raw_opssight, onlyHub, both, onlyOpssight):
    print("==== ERROR LOG ====")
    for elem in onlyHub:
        if elem in raw_opssight or elem not in raw_hub:
            print("ERROR - onlyHub "+elem)      
    for elem in both:
        if elem not in raw_hub or elem not in raw_opssight:
            print("ERROR - both "+elem)
    for elem in onlyOpssight:
        if elem not in raw_opssight or elem in raw_hub:
            print("ERROR - onlyOpsight {}".format(elem))

def tests(k8s_client, hub_client, opssight_client, PORT_NUMBER):
    global STATUS
    if len(hub_client.get_projects_names()) > 0 and len(opssight_client.get_shas_names()) > 0:
        print("Connected to Hub!")
        print("Connected to OpsSight!")
        sys.stdout.flush()
        STATUS = b'PASSED'
    else:
        print("Couldn't Connect to Hub")
        sys.stdout.flush()
        STATUS = b'FAILED'

    try:
        server = HTTPServer(('', PORT_NUMBER), myHandler)
        server.serve_forever()
    except:
        server.socket.close()
    

def main():
    '''logger = logging.getLogger('output')
    logger.setLevel(logging.INFO)
    
    ch = logging.StreamHandler(sys.stdout)
    ch.setLevel(logging.INFO)'''

    if len(sys.argv) < 1:
        print("Invalid Parameters")
        return 0
    # path to config
    test_config_path = sys.argv[1]
    test_config_json = None
    with open(test_config_path) as f:
        test_config_json = json.load(f)
    opssight_url = test_config_json["PerceptorURL"]
    hub_url = test_config_json["HubURL"]
    PORT_NUMBER = test_config_json["Port"]
    usr = test_config_json["Username"]
    password = test_config_json["Password"]

    # Create Clients to access Cluster Data
    k8s_client = None #k8sClientWrapper()
    hub_client = HubClient(hub_url, None, usr, password)
    opssight_client = OpsSightClient(opssight_url, None)

    # Check if Scans are being performed
    '''logger.info("Running Tests")'''
    sys.stdout.flush()
    #tests(k8s_client, hub_client, opssight_client, PORT_NUMBER)
    print()

    hub_IDs = hub_client.get_code_locations_names()
    '''print("Scans in the Hub: ",len(hub_IDs))
    print(sorted(hub_IDs))
    print()'''

    opssight_IDs = opssight_client.get_shas_names()
    '''print("Scans in OpsSight: ",len(opssight_IDs))
    print(sorted(opssight_IDs))
    print()'''

    IDs_hub_only, IDs_hub_and_opssight, IDs_opssight_only = diff_hub_IDs_and_opssight_IDs(hub_IDs, opssight_IDs)
    '''print("Scans Only on OpSight: ",len(IDs_opssight_only))
    print(sorted(IDs_opssight_only))
    print()'''

    '''print("Scans Only on Hub: ", len(IDs_hub_only))
    print(sorted(IDs_hub_only))
    print()

    print("Scans Found in Both: ", len(IDs_hub_and_opssight))
    print(sorted(IDs_hub_and_opssight))
    print()'''

    print("")
    print("Images In Your Kubernetes Cluster : ", len(IDs_opssight_only)+len(IDs_hub_only))
    print("Total Images Scanned in Hub1 : ", len(hub_IDs))
    print("Total Images Scanned in Hub2 : ", 0)
    print("Cumulative Images Scanned : ", len(opssight_IDs))
    print("\n")
    print("***************************************")
    print("OpsSight Test Threshold : ",100,"% Image coverage to ship.")
    coverage = len(IDs_opssight_only)+len(IDs_hub_only) 
    print("OpsSight Target Completion : ",coverage  ,"% Image coverage")
    print("***************************************")

    if coverage < 100 :
        print("OpsSight Test Result: No Automated Release is Possible at this time.")
    else:
        print("OpsSight Test Result: PASS")
    
    print("***************************************")

    return 0


main()