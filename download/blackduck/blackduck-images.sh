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

BLACKDUCK_IMAGES=(
"$IMAGE_PREFIX"-authentication:"$_arg_tag" \
"$IMAGE_PREFIX"-documentation:"$_arg_tag" \
"$IMAGE_PREFIX"-jobrunner:"$_arg_tag" \
"$IMAGE_PREFIX"-registration:"$_arg_tag" \
"$IMAGE_PREFIX"-scan:"$_arg_tag" \
"$IMAGE_PREFIX"-webapp:"$_arg_tag" \
"$IMAGE_PREFIX"-cfssl:"$_arg_cfssl" \
"$IMAGE_PREFIX"-logstash:"$_arg_logstash" \
"$IMAGE_PREFIX"-nginx:"$_arg_nginx" \
"$IMAGE_PREFIX"-solr:"$_arg_solr" \
"$IMAGE_PREFIX"-zookeeper:"$_arg_zookeeper" \
)

# If binary scanner is enabled, add those images to the array
if [[ ! "$_arg_binary_scanner" == "off" ]]; then
    BLACKDUCK_IMAGES+=(
    "appcheck-worker":"$_arg_binaryscanner" \
    "rabbitmq":"$_arg_rabbitmq" \
    "blackduck-upload-cache":"$_arg_uploadcache" \
    )
fi

echo "*************************************************************************"
echo "Started pulling all Black Duck images"
echo "*************************************************************************"
# To support black duck images
for BLACKDUCK_IMAGE in "${BLACKDUCK_IMAGES[@]}"
do
    KEY="${BLACKDUCK_IMAGE%%:*}"
    VALUE="${BLACKDUCK_IMAGE##*:}"
    docker pull "docker.io/blackducksoftware/$KEY:$VALUE"; rc=$?;
    if [[ $rc != 0 ]]; then
        echo "Unable to pull the image because version of the image might not exist or doesn't have an access to pull the image"
        exit 1
    fi
done
# To support external postgres images
docker pull "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1"; rc=$?;
if [[ $rc != 0 ]]; then
    echo "Unable to pull the image because version of the image might not exist or doesn't have an access to pull the image"
    exit 1
fi
echo "*************************************************************************"
echo "Pulled all Black Duck images"
echo "*************************************************************************"
echo ""
echo ""

BLACKDUCK_DIR="blackduck-images"
if [[ "$_arg_push"  == "off" ]]; then
    mkdir -p ./"$BLACKDUCK_DIR"
    echo "*************************************************************************"
    echo "Started saving all Black Duck images in ./$BLACKDUCK_DIR directory. Please wait!!!"
    echo "*************************************************************************"
    for BLACKDUCK_IMAGE in "${BLACKDUCK_IMAGES[@]}"
    do
        KEY="${BLACKDUCK_IMAGE%%:*}"
        VALUE="${BLACKDUCK_IMAGE##*:}"
        docker save "docker.io/blackducksoftware/$KEY:$VALUE" -o ./"$BLACKDUCK_DIR/$KEY.tar"
    done
    docker save "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1" -o ./"$BLACKDUCK_DIR/postgresql-96-rhel7.tar"
    echo "*************************************************************************"
    echo "Saved all Black Duck images"
    echo "*************************************************************************"
else
    echo "********************************************************************************************************************"
    echo "Please provide the Docker credentials of $_arg_registry registry for $_arg_user user..."
    echo "********************************************************************************************************************"
    docker login -u "$_arg_user" "$_arg_registry"
    echo ""
    echo ""

    DOCKER_REPO="$_arg_registry"/"$_arg_project"

    # Docker tag all Black Duck images
    echo "*************************************************************************"
    echo "Started tagging and pushing all Black Duck images"
    echo "*************************************************************************"
    for BLACKDUCK_IMAGE in "${BLACKDUCK_IMAGES[@]}"
    do
        KEY="${BLACKDUCK_IMAGE%%:*}"
        VALUE="${BLACKDUCK_IMAGE##*:}"
        docker tag "docker.io/blackducksoftware/$KEY:$VALUE" "$DOCKER_REPO/$KEY:$VALUE"
        docker push "$DOCKER_REPO/$KEY:$VALUE"; rc=$?;
        if [[ $rc != 0 ]]; then
            echo "Unable to push the image because Docker login failed or doesn't have an access to push the image"
            exit 1
        fi
    done
    docker tag "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1" "$DOCKER_REPO/postgresql-96-rhel7:1"
    docker push "$DOCKER_REPO/postgresql-96-rhel7:1"; rc=$?;
    if [[ $rc != 0 ]]; then
        echo "Unable to push the postgres image because Docker login failed or doesn't have an access to push the image"
        exit 1
    fi
    echo "*************************************************************************"
    echo "Tagged and pushed all Black Duck images"
    echo "*************************************************************************"
fi