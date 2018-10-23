import sys 
import os
import json
import subprocess

NUM_MAX_PROJECTS=10000000

''' 
K8S Client 
'''

def get_k8s_namespaces():
    # K8s command to get a table of namespaces without the header line
    ns, err = subprocess.Popen(["oc","get", "ns","--no-headers"], stdout=subprocess.PIPE).communicate()
    # Split output by row
    ns_list = ns.split("\n")[:-1] # remove empty string as last entry
    # Get first entry in each row
    ns_list = [n.split()[0] for n in ns_list]
    return ns_list
    

def get_k8s_images(namespace=""):
    # Select all namespaces if one wasn't provided
    if namespace == '':
        namespace = '--all-namespaces'
    else:
        namespace = '-n '+namespace 
    # Path to an image within a Pod's spec
    image_path = '{.items[*].spec.containers[*].image}'
    # K8s Terminal Command to get all pods and extract the images from the pod's json
    k8s_command = ''.join(['oc get pods ',namespace,' -o jsonpath="',image_path,'"'])
    images, err = subprocess.Popen(k8s_command, shell=True, stdout=subprocess.PIPE).communicate()
    # Remove initial ", split output, and remove last entry (vault:latest")
    return images[1:].split()

def get_images_per_pod():
    pods = json.loads(subprocess.Popen("oc get pods --all-namespaces -o json", shell=True, stdout=subprocess.PIPE).communicate()[0])
    pod_dump = []
    for pod in pods['items']:
        ns = pod['metadata']['namespace']
        try:
            for container in pod['status']['containerStatuses']:
                name = container['name']
                image = container['image']
                imageID = container['imageID']
                pod_dump.append([ns, name, image, imageID])
        except:
            print("======== NO CONTAINER STATUSES ========")
            print(json.dumps(pod, indent=2))
    return pod_dump


def get_annotations_per_pod():
    pass 

def get_labels_per_pod():
    pass


def get_all_projects(): 
    # K8s command to get all projects
    projects, err = subprocess.Popen(["oc","get", "projects", "--no-headers"], stdout=subprocess.PIPE).communicate()
    projects_list = ns.split("\n")[:-1] # remove empty string as last entry
    ns_list = [n.split()[0] for n in ns_list]
    # split output and remove last entry (vault:latest)
    return projects.split()[:-1] 

'''
Hub Client
'''

def get_hub_projects_dump(host_name): 
    secure_login_cookie = get_secure_login_cookie(host_name)
    projects_json, err = subprocess.Popen("curl --insecure -sX GET -H 'Accept: application/json' -H 'Cookie: "+secure_login_cookie+"' https://"+host_name+":443/api/projects?limit="+str(NUM_MAX_PROJECTS), shell=True, stdout=subprocess.PIPE).communicate()
    d = json.loads(projects_json)
    return d['items']

def get_hub_projects_names(host_name):
    return [x['name'] for x in get_hub_projects_dump(host_name)]


def get_hub_code_locations_dump(host_name):
    secure_login_cookie = get_secure_login_cookie(host_name)
    code_locs_json, err = subprocess.Popen("curl --insecure -sX GET -H 'Accept: application/json' -H 'Cookie: "+secure_login_cookie+"' https://"+host_name+":443/api/codelocations?limit="+str(NUM_MAX_PROJECTS), shell=True, stdout=subprocess.PIPE).communicate()
    d = json.loads(code_locs_json)
    return d

def get_hub_code_locations_names(host_name):
    return [x['name'] for x in get_hub_code_locations_dump(host_name)['items']]

'''
OpsSight Client
'''

def get_opssight_dump():
    code_locs_json, err = subprocess.Popen("curl -sX GET -H 'Accept: application/json' http://perceptor-ops.10.1.176.68.xip.io/model", shell=True, stdout=subprocess.PIPE).communicate()
    d = json.loads(code_locs_json)
    return d

def get_opssight_shas_names():
    return get_opssight_dump()['CoreModel']['Images'].keys()


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


def namespace_test():
    initial_namespaces = set(get_k8s_namespaces())
    ns, err = subprocess.Popen(["oc", "new-project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    updated_namespaces = set(get_k8s_namespaces())
    ns, err = subprocess.Popen(["oc", "delete", "project", "opssight-smoke-test"], stdout=subprocess.PIPE).communicate()
    final_namespaces = set(get_k8s_namespaces())

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
Tools
'''

def get_secure_login_cookie(host_name):
    secure_login_cookie, err = subprocess.Popen("curl -si --insecure --header 'Content-Type: application/x-www-form-urlencoded' --request POST --data 'j_username=sysadmin&j_password=blackduck' https://"+host_name+":443/j_spring_security_check | awk '/Set-Cookie/{print $2}' | sed 's/;$//'", shell=True, stdout=subprocess.PIPE).communicate()
    return secure_login_cookie

'''
main
'''

def check_namespaces_loop():
    old_namespaces = set(get_k8s_namespaces())
    while True:
        subprocess.Popen("sleep 5", shell=True, stdout=subprocess.PIPE).communicate()
        new_namespaces = set(get_k8s_namespaces())
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
    #print(get_k8s_namespaces())
    #print(get_k8s_images())
    #print(get_hub_projects_names("aci-471-aci-471.10.1.176.130.xip.io"))
    
    #hub_IDs = get_hub_code_locations_names("aci-471-aci-471.10.1.176.130.xip.io")
    #opssight_IDs = get_opssight_shas_names()
    #IDs_hub_only, IDs_hub_and_opssight, IDs_opssight_only = diff_hub_IDs_and_opssight_IDs(hub_IDs, opssight_IDs)
    #test(hub_IDs, opssight_IDs, IDs_hub_only, IDs_hub_and_opssight, IDs_opssight_only)

    #namespace_test()

    #get_images_per_pod()

    #print(get_k8s_images("mphammer"))
    check_namespaces_loop()


main()