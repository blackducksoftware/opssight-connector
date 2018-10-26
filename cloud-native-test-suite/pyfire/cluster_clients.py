import json
import subprocess
import requests
import time
import sys
from kubernetes import client, config
from http.server import BaseHTTPRequestHandler, HTTPServer
import logging 


NUM_MAX_PROJECTS=10000000

class myHandler(BaseHTTPRequestHandler):
    def __init__(self):
        self.status = "UNKNOWN"

    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type','text/html')
        self.end_headers()
        self.wfile.write(self.status)

    def my_own_server_function():
        # TO DO
        pass


class K8sClient:
    def __init__(self):
        config.load_kube_config()
        self.v1 = client.CoreV1Api()

    def get_pods_kube(self):
        ret = self.v1.list_pod_for_all_namespaces(watch=False)
        for i in ret.items:
            print("%s\t%s\t%s" % (i.status.pod_ip, i.metadata.namespace, i.metadata.name))

    def get_api_resources_kube(self):
        resources = self.v1.get_api_resources()
        return [x.name for x in resources.resources]
    
    def get_namespaces_kube(self):
        names = []
        for ns in self.v1.list_namespace().items:
            names.append(ns.metadata.name)
        return names

    def get_namespaces(self):
        # K8s command to get a table of namespaces without the header line
        ns, err = subprocess.Popen(["oc","get", "ns","--no-headers"], stdout=subprocess.PIPE).communicate()
        # Split output by row
        ns_list = ns.decode('unicode_escape').split("\n")[:-1] # remove empty string as last entry
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


class HubClient:
    def __init__(self, host_name=None, k8s=None, username="", password="", yaml_path="hub.yml"):
        self.host_name = host_name
        self.username = username
        self.password = password 
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
        # todo fix
        security_data = {'j_username': self.username,'j_password': self.password}
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


class OpsSightClient:
    def __init__(self, host_name=None, k8s=None, yaml_path="opssight.yml"):
        self.host_name = host_name
        self.yaml_path = yaml_path
        self.k8s = k8s

    def create(self):
        self.create_yaml()
        
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

    def create_in_cluster(self):
        print("Creating Yaml File")
        self.create_yaml()
        
        self.host_name = ns.split()[1].decode('unicode_escape')

    def create_yaml(self):
        ns, err = subprocess.Popen(["./create-opssight-yaml.sh"], stdout=subprocess.PIPE).communicate()
    
    def destroy(self):
        ns, err = subprocess.Popen(["oc", "delete", "-f", self.yaml_path], stdout=subprocess.PIPE).communicate()
    
    def get_dump(self):
        while True:
            r = requests.get("http://"+self.host_name+"/model")
            #print(r.text)
            #if 200 <= r.status_code < 300:
            if r.status_code == 200:
                return json.loads(r.text)
        

    def get_shas_names(self):
        return self.get_dump()['CoreModel']['Images'].keys()