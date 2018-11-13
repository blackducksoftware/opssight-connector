#!/bin/bash

source `dirname ${BASH_SOURCE}`/args.sh "${@}"

if [[ "$_arg_push"  == "on" && "$_arg_project"  == "" ]]; then
    echo "Please provide the Docker project/repository to push the images!!!"
    exit 1
fi

OPSSIGHT_IMAGES=("blackduck-operator" \
"opssight-core" \
"opssight-scanner" \
"opssight-image-getter" \
"opssight-pod-processor" \
"opssight-image-processor")

echo "*************************************************************************"
echo "Started pulling all OpsSight images"
echo "*************************************************************************"
for OPSSIGHT_IMAGE in "${OPSSIGHT_IMAGES[@]}"
do
	docker pull docker.io/blackducksoftware/"$OPSSIGHT_IMAGE":"$_arg_tag"
done
echo "*************************************************************************"
echo "Pulled all OpsSight images"
echo "*************************************************************************"

OPSSIGHT_DIR="opssight-images"
if [[ "$_arg_push"  == "off" ]]; then
    mkdir -p ./"$OPSSIGHT_DIR"
    echo "*************************************************************************"
    echo "Started saving all OpsSight images"
    echo "*************************************************************************"
    for OPSSIGHT_IMAGE in "${OPSSIGHT_IMAGES[@]}"
    do
        docker save docker.io/blackducksoftware/"$OPSSIGHT_IMAGE":"$_arg_tag" -o ./"$OPSSIGHT_DIR"/"$OPSSIGHT_IMAGE".tar
    done
    echo "*************************************************************************"
    echo "Saved all OpsSight images in ./$OPSSIGHT_DIR"
    echo "*************************************************************************"
else
    echo ""
    echo ""
    echo "********************************************************************************************************************"
    echo "Please provide the Docker credentials of $_arg_registry registry for $_arg_user user..."
    echo "********************************************************************************************************************"
    docker login -u "$_arg_user" "$_arg_registry"

    # Docker tag and push all opssight images
    DOCKER_REPO="$_arg_registry"/"$_arg_project"
    echo "*************************************************************************"
    echo "Started tagging and pushing all OpsSight images"
    echo "*************************************************************************"
    for OPSSIGHT_IMAGE in "${OPSSIGHT_IMAGES[@]}"
    do
        docker tag docker.io/blackducksoftware/"$OPSSIGHT_IMAGE":"$_arg_tag" "$DOCKER_REPO"/"$OPSSIGHT_IMAGE":"$_arg_tag"
        docker push "$DOCKER_REPO"/"$OPSSIGHT_IMAGE":"$_arg_tag"
    done
    echo "*************************************************************************"
    echo "Tagged and pushed all OpsSight images"
    echo "*************************************************************************"
fi