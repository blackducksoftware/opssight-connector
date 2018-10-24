import json
import subprocess
import requests
import time
import sys

NUM_MAX_PROJECTS=10000000

class Client:
    def k8sCommand():
        pass

''' 
K8S Client 
'''
class K8sClient:
    def __init__(self):
        pass

    def get_namespaces(self):
        # K8s command to get a table of namespaces without the header line
        ns, err = subprocess.Popen(["oc","get", "ns","--no-headers"], stdout=subprocess.PIPE).communicate()
        # Split output by row
        ns_list = ns.split("\n")[:-1] # remove empty string as last entry
        # Get first entry in each row
        ns_list = [n.split()[0] for n in ns_list]
        return ns_list
        

    def get_images(self, namespace=""):
        # Select all namespaces if one wasn't provided
        namespace = '--all-namespaces' if namespace == '' else '-n '+namespace 
        # Path to an image within a Pod's spec
        image_path = '{.items[*].spec.containers[*].image}'
        # K8s Terminal Command to get all pods and extract the images from the pod's json
        k8s_command = ''.join(['oc get pods ',namespace,' -o jsonpath="',image_path,'"'])
        images, err = subprocess.Popen(k8s_command, shell=True, stdout=subprocess.PIPE).communicate()
        # Remove initial "
        return images[1:].split()

    def get_images_per_pod(self, namespace=""):
        namespace = '--all-namespaces' if namespace == '' else '-n '+namespace 
        k8s_command = ''.join(["oc get pods ",namespace," -o json"])
        pods = json.loads(subprocess.Popen(k8s_command, shell=True, stdout=subprocess.PIPE).communicate()[0])
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

    def get_annotations_per_pod(self, namespace=""):
        namespace = '--all-namespaces' if namespace == '' else '-n '+namespace 
        k8s_command = ''.join(["oc get pods ",namespace," -o json"])
        pods = json.loads(subprocess.Popen(k8s_command, shell=True, stdout=subprocess.PIPE).communicate()[0])
        pod_dump = []
        for pod in pods['items']:
            ns = pod['metadata']['namespace']
            name = pod['metadata']['name']
            annotation = pod['metadata']['annotations']
            pod_dump.append([ns, name, annotation])
        return pod_dump

    def get_labels_per_pod(self, namespace=""):
        namespace = '--all-namespaces' if namespace == '' else '-n '+namespace 
        k8s_command = ''.join(["oc get pods ",namespace," -o json"])
        pods = json.loads(subprocess.Popen(k8s_command, shell=True, stdout=subprocess.PIPE).communicate()[0])
        pod_dump = []
        for pod in pods['items']:
            ns = pod['metadata']['namespace']
            name = pod['metadata']['name']
            annotation = pod['metadata']['labels']
            pod_dump.append([ns, name, annotation])
        return pod_dump

    def get_all_projects(self): 
        # K8s command to get all projects
        projects, err = subprocess.Popen(["oc","get", "projects", "--no-headers"], stdout=subprocess.PIPE).communicate()
        projects_list = ns.split("\n")[:-1] # remove empty string as last entry
        ns_list = [n.split()[0] for n in ns_list]
        return projects.split()

'''
Hub Client
'''
class HubClient:
    def __init__(self, host_name, yaml_path="hub.yml"):
        self.host_name = host_name
        self.secure_login_cookie = self.get_secure_login_cookie()
        self.yaml_path = yaml_path

    def create(self):
        self.create_yaml()
        ns, err = subprocess.Popen(["oc", "create", "-f", self.yaml_path], stdout=subprocess.PIPE).communicate()

    def create_yaml(self):
        ns, err = subprocess.Popen(["./create-hub-yaml.sh"], stdout=subprocess.PIPE).communicate()

    def destory(self):
         ns, err = subprocess.Popen(["oc", "delete", "-f", self.yaml_path], stdout=subprocess.PIPE).communicate()

    def get_secure_login_cookie(self):
        security_headers = {'Content-Type':'application/x-www-form-urlencoded'}
        security_data = {'j_username':'sysadmin','j_password':'blackduck'}
        # verify=False does not verify SSL connection - insecure
        r = requests.post("https://"+self.host_name+":443/j_spring_security_check", verify=False, data=security_data, headers=security_headers)
        return r.cookies 
        
    def get_projects_dump(self): 
        r = requests.get("https://"+self.host_name+":443/api/projects?limit="+str(NUM_MAX_PROJECTS),verify=False, cookies=self.secure_login_cookie)
        return r.json()['items']

    def get_projects_names(self):
        return [x['name'] for x in self.get_projects_dump()]

    def get_code_locations_dump(self):
        r = requests.get("https://"+self.host_name+":443/api/codelocations?limit="+str(NUM_MAX_PROJECTS),verify=False, cookies=self.secure_login_cookie)
        return r.json()

    def get_code_locations_names(self):
        return [x['name'] for x in self.get_code_locations_dump()['items']]

'''
OpsSight Client
'''

class OpsSightClient:
    def __init__(self, host_name=None, yaml_path="opssight.yml"):
        self.host_name = host_name
        self.yaml_path = yaml_path

    def create(self):
        #self.create_yaml()
        
        print("Pushing OpsSight Yaml.")
        ns, err = subprocess.Popen(["oc", "create", "-f", self.yaml_path], stdout=subprocess.PIPE).communicate()
        print("")

        print("Exposing Perceptor Service.",end="")
        ns, err = subprocess.Popen(["oc", "expose", "svc", "perceptor", "-n", "ops"], stdout=subprocess.PIPE,stderr=subprocess.PIPE).communicate()
        while err != b'':
            ns, err = subprocess.Popen(["oc", "expose", "svc", "perceptor", "-n", "ops"], stdout=subprocess.PIPE,stderr=subprocess.PIPE).communicate()
            print(".",end="")
            sys.stdout.flush()
        print("\n")
        sys.stdout.flush()

        print("Getting the route to connect to.",end="")
        ns, err = subprocess.Popen(["oc", "get", "routes", "-n", "ops", "--no-headers"], stdout=subprocess.PIPE,stderr=subprocess.PIPE).communicate()
        while err != b'':
            ns, err = subprocess.Popen(["oc", "get", "routes", "-n", "ops", "--no-headers"], stdout=subprocess.PIPE,stderr=subprocess.PIPE).communicate()
            print(".",end="")
            sys.stdout.flush()
        print("\n")
        sys.stdout.flush()
        print("Route: "+str(ns.split()[1].decode('unicode_escape')))
        
        self.host_name = ns.split()[1].decode('unicode_escape')

    def create_yaml(self):
        ns, err = subprocess.Popen(["./create-opssight-yaml.sh"], stdout=subprocess.PIPE).communicate()
    
    def destroy(self):
        print("Tearing down OpsSight Yaml")
        ns, err = subprocess.Popen(["oc", "delete", "-f", self.yaml_path], stdout=subprocess.PIPE).communicate()
    
    def get_dump(self):
        while True:
            r = requests.get("http://"+self.host_name+"/model")
            if 200 <= r.status_code < 300:
                return json.loads(r.text)
        

    def get_shas_names(self):
        return self.get_dump()['CoreModel']['Images'].keys()