import sys 
import os
import json
import subprocess
from cluster_clients import *


'''
Tests
'''

def diff_hub_IDs_and_opssight_IDs(hub_IDs, opssight_IDs):
    hub_IDs_set = set(hub_IDs)
    opssight_IDs_set = set(opssight_IDs)
    
    IDs_hub_only = hub_IDs_set.difference(opssight_IDs_set)
    IDs_both = hub_IDs_set.intersection(opssight_IDs_set)
    IDs_opssight_only = opssight_IDs_set.difference(hub_IDs_set)

    return IDs_hub_only, IDs_both, IDs_opssight_only
    
def test(raw_hub, raw_opssight, onlyHub, both, onlyOpssight):
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


def create_opssight_test():
    ns, err = subprocess.Popen(["oc", "new-project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    create_opsight_yaml("300m")
    ns, err = subprocess.Popen(["oc", "create", "-f", "create-opssight.yml"], stdout=subprocess.PIPE).communicate()
    # Python Script Testing
    if err == 0:
        print("BAD!")
    else:
        print("GOOD")
    ns, err = subprocess.Popen(["oc", "delete", "project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()

def create_opsight_yaml(arg1=""):
    ns, err = subprocess.Popen(["./create-opsight.sh",arg1], stdout=subprocess.PIPE).communicate()


def namespace_test(k8s):
    initial_namespaces = set(k8s.get_namespaces())
    ns, err = subprocess.Popen(["oc", "new-project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    updated_namespaces = set(k8s.get_namespaces())
    ns, err = subprocess.Popen(["oc", "delete", "project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    final_namespaces = set(k8s.get_namespaces())

    new_namespaces = list(updated_namespaces.difference(initial_namespaces))
    
    if len(new_namespaces) == 1:
        print("New Namespace: "+new_namespaces[0])
    elif len(new_namespaces) > 1:
        print("ERROR - Multiple Namespaces were created")
        print(new_namespaces)
    else:
        print("ERROR - No new Namespaces were created")

    new_namespaces = list(final_namespaces.difference(initial_namespaces))
    if len(new_namespaces) > 0:
        print("ERROR - New Namespaces are in existence")
        print(new_namespaces)
    else:
        print("Namespaces Restored")

'''
main
'''

def check_namespaces_loop(k8s):
    old_namespaces = set(k8s.get_namespaces())
    while True:
        subprocess.Popen("sleep 5", shell=True, stdout=subprocess.PIPE).communicate()
        new_namespaces = set(k8s.get_namespaces())
        added_namespaces = new_namespaces.difference(old_namespaces)
        removed_namespaces = old_namespaces.difference(new_namespaces)
        if len(added_namespaces) !=0 or len(removed_namespaces) != 0:
            if len(added_namespaces) > 0:
                print("Added Namespaces:")
                print(added_namespaces)
            if len(removed_namespaces) > 0:
                print("Removed Namespaces:")
                print(removed_namespaces)
        else: 
            print("No changes to Namespaces")
        old_namespaces=new_namespaces

def main():
    k8s = K8sClient()
    #hub = HubClient("aci-471-aci-471.10.1.176.130.xip.io")
    hub = HubClient('engsreepath471-engsreepath471.10.1.176.130.xip.io')
    #opssight = OpsSightClient('perceptor-ops.10.1.176.68.xip.io')
    #print(k8s.get_namespaces())
    #print(k8s.get_images())
    #print(hub.get_projects_names())
    
    #hub_IDs = hub.get_code_locations_names()
    #opssight_IDs = opssight.get_shas_names()
    #IDs_hub_only, IDs_hub_and_opssight, IDs_opssight_only = diff_hub_IDs_and_opssight_IDs(hub_IDs, opssight_IDs)
    #test(hub_IDs, opssight_IDs, IDs_hub_only, IDs_hub_and_opssight, IDs_opssight_only)

    #k8s.get_images_per_pod()

    #print(k8s.get_images("mphammer"))
    
    #namespace_test(k8s)

    #check_namespaces_loop(k8s)

    print(k8s.get_annotations_per_pod())


main()