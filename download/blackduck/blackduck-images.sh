#!/bin/bash

source `dirname ${BASH_SOURCE}`/args.sh "${@}"

if [[ "$_arg_push"  == "on" && "$_arg_project"  == "" ]]; then
    echo "Please provide the Docker project/repository to push the images!!!"
    exit 1
fi

IMAGE_PREFIX="blackduck"
if [[ "$_arg_tag" == "4."* ]] || [[ "$_arg_tag" == "5.0"* ]]; then
    echo "Image prefix is hub"
    IMAGE_PREFIX="hub"
fi

BLACKDUCK_IMAGES=("$IMAGE_PREFIX"-cfssl \
"$IMAGE_PREFIX"-postgres \
"$IMAGE_PREFIX"-jobrunner \
"$IMAGE_PREFIX"-nginx \
"$IMAGE_PREFIX"-webapp \
"$IMAGE_PREFIX"-logstash \
"$IMAGE_PREFIX"-documentation \
"$IMAGE_PREFIX"-solr \
"$IMAGE_PREFIX"-registration \
"$IMAGE_PREFIX"-zookeeper \
"$IMAGE_PREFIX"-authentication)

echo "*************************************************************************"
echo "Started pulling all Black Duck images"
echo "*************************************************************************"
for BLACKDUCK_IMAGE in "${BLACKDUCK_IMAGES[@]}"
do
    docker pull docker.io/blackducksoftware/"$BLACKDUCK_IMAGE":"$_arg_tag"
done
docker pull registry.access.redhat.com/rhscl/postgresql-96-rhel7:1
echo "*************************************************************************"
echo "Pulled all Black Duck images"
echo "*************************************************************************"

BLACKDUCK_DIR="blackduck-images"
if [[ "$_arg_push"  == "off" ]]; then
    mkdir -p ./"$BLACKDUCK_DIR"
    echo "*************************************************************************"
    echo "Started saving all Black Duck images"
    echo "*************************************************************************"
    for BLACKDUCK_IMAGE in "${BLACKDUCK_IMAGES[@]}"
    do
        docker save docker.io/blackducksoftware/"$BLACKDUCK_IMAGE":"$_arg_tag" -o ./"$BLACKDUCK_DIR"/"$BLACKDUCK_IMAGE".tar
    done
    docker save registry.access.redhat.com/rhscl/postgresql-96-rhel7:1 -o ./"$BLACKDUCK_DIR"/postgresql-96-rhel7.tar
    echo "*************************************************************************"
    echo "Saved all Black Duck images"
    echo "*************************************************************************"
else
    echo ""
    echo ""
    echo "********************************************************************************************************************"
    echo "Please provide the Docker credentials of $_arg_registry registry for $_arg_user user..."
    echo "********************************************************************************************************************"
    docker login -u "$_arg_user" "$_arg_registry"

    DOCKER_REPO="$_arg_registry"/"$_arg_project"

    # Docker tag all Black Duck images
    echo "*************************************************************************"
    echo "Started tagging and pushing all Black Duck images"
    echo "*************************************************************************"
    for BLACKDUCK_IMAGE in "${BLACKDUCK_IMAGES[@]}"
    do
        docker tag docker.io/blackducksoftware/"$BLACKDUCK_IMAGE":"$_arg_tag" "$DOCKER_REPO"/"$BLACKDUCK_IMAGE":"$_arg_tag"
        docker push "$DOCKER_REPO"/"$BLACKDUCK_IMAGE":"$_arg_tag"
    done
    docker tag registry.access.redhat.com/rhscl/postgresql-96-rhel7:1 "$DOCKER_REPO"/postgresql-96-rhel7:1
    docker push "$DOCKER_REPO"/postgresql-96-rhel7:1
    echo "*************************************************************************"
    echo "Tagged and pushed all Black Duck images"
    echo "*************************************************************************"
fi