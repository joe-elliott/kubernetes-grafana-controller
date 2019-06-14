export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false

GRAFANA_URL=""
CONTROLLER_URL=""

validateGrafanaUrl() {

    run kubectl wait pod -l app=grafana --for condition=ready --timeout=30s

    # ugh
    sleep 5s

    # urls
    GRAFANA_URL=$(minikube service grafana --url --interval=1 --wait=60)

    echo "grafana url: " $GRAFANA_URL
    httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL})

    [ "$httpStatus" -eq "200" ]
}

validateControllerUrl() {
    CONTROLLER_URL=$(minikube service kubernetes-grafana-test --url)

    httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${CONTROLLER_URL}/metrics)

    [ "$httpStatus" -eq "200" ]
}

#
# validateFolderCount <count>
#   use grafana folder api to confirm that the count is what is expected 
#
validateFolderCount() {
    folderJson=$(curl --silent ${GRAFANA_URL}/api/folders)

    count=$(echo $folderJson | jq length)

    if [ "$count" -ne "$1" ]; then
        echo "count: $count param: $1"
    fi

    [ "$count" -eq "$1" ]
}

#
# validatePostFolder <yaml file name>
#   note that the dashboard file name must match the Dashboard object name ...
#    ... b/c i'm lazy
#
validatePostFolder() {
    specfile=$1

    folderName=$(objectNameFromFile $specfile)

    # create in kubernetes
    kubectl apply -f $specfile >&2

	sleep 5s

    folderId=$(kubectl get Folder -o=jsonpath="{.items[?(@.metadata.name==\"${folderName}\")].status.grafanaID}")

    # check if exists in grafana
	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/folders/${dashboardId})
    [ "$httpStatus" -eq "200" ]

    echo $folderId
}

#
# validateDashboardCount <count>
#   use grafana search api to confirm that the count is what is expected 
#
validateDashboardCount() {
    searchJson=$(curl --silent ${GRAFANA_URL}/api/search)

    count=$(echo $searchJson | jq length)

    if [ "$count" -ne "$1" ]; then
        echo "count: $count param: $1"
    fi

    [ "$count" -eq "$1" ]
}

#
# validatePostDashboard <yaml file name>
#   note that the dashboard file name must match the Dashboard object name ...
#    ... b/c i'm lazy
#
validatePostDashboard() {
    specfile=$1

    dashboardName=$(objectNameFromFile $specfile)

    # create in kubernetes
    kubectl apply -f $specfile >&2

	sleep 5s

    dashboardId=$(kubectl get Dashboard -o=jsonpath="{.items[?(@.metadata.name==\"${dashboardName}\")].status.grafanaID}")

    # check if exists in grafana
	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
    [ "$httpStatus" -eq "200" ]

    echo $dashboardId
}

#
# validateDashboardContents <yaml file name>
#   creates a dashboard and both verifies it exists and that its content matches
#
validateDashboardContents() {
    filename=$1

    dashboardId=$(validatePostDashboard $filename)

    echo "Test Json Content of $filename ($dashboardId)"

    dashboardJsonFromGrafana=$(curl --silent ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})

    echo $dashboardJsonFromGrafana | jq '.dashboard | del(.version) | del(.id) | del (.uid)' > a.json

    dashboardJsonFromYaml=$(grep -A9999 'json' $filename)
    dashboardJsonFromYaml=${dashboardJsonFromYaml%?}   # strip final quote
    dashboardJsonFromYaml=${dashboardJsonFromYaml#*\'} # strip up to and including the first quote

    echo $dashboardJsonFromYaml | jq 'del(.version) | del(.id) | del (.uid)' > b.json

    equal=$(jq --argfile a a.json --argfile b b.json -n '$a == $b')

    if [ "$equal" != "true" ]; then
        run diff <(jq -S . a.json) <(jq -S . b.json)
        echo $output
    fi

    [ "$equal" = "true" ]

    rm a.json
    rm b.json
}

#
# alert notifications
#

# validatePostAlertNotification <yaml file name>
#   note that the notification file name must match the AlertNotification object name ...
#    ... b/c i'm lazy
#
validatePostAlertNotification() {
    specfile=$1

    notificationName=$(objectNameFromFile $specfile)

    # create in kubernetes
    kubectl apply -f $specfile >&2

	sleep 5s

    notificationId=$(kubectl get AlertNotification -o=jsonpath="{.items[?(@.metadata.name==\"${notificationName}\")].status.grafanaID}")

    [ "$notificationId" != "" ]

    # check if exists in grafana
	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/alert-notifications/${notificationId})
    [ "$httpStatus" -eq "200" ]

    echo $notificationId
}


#
# validateAlertNotificationCount <count>
#   use grafana search api to confirm that the count is what is expected 
#
validateAlertNotificationCount() {
    notificationJson=$(curl --silent ${GRAFANA_URL}/api/alert-notifications)

    count=$(echo $notificationJson | jq length)

    [ "$count" -eq "$1" ]
}

#
# validateAlertNotificationContents <yaml file name>
#   creates a AlertNotification and both verifies it exists and that its content matches
#
validateAlertNotificationContents() {
    filename=$1

    notificationId=$(validatePostAlertNotification $filename)

    echo "Test Json Content of $filename ($notificationId)"

    notificationJsonFromYaml=$(grep -A9999 'json' $filename)
    notificationJsonFromYaml=${notificationJsonFromYaml%?}   # strip final quote
    notificationJsonFromYaml=${notificationJsonFromYaml#*\'} # strip up to and including the first quote

    echo $notificationJsonFromYaml > b.json
    fieldsToKeep=$(cat b.json | jq keys)

    notificationJsonFromGrafana=$(curl --silent ${GRAFANA_URL}/api/alert-notifications/${notificationId})

    # grafana can add fields to flesh out the object.  remove anything from grafana not present in the original
    #  spec file
    echo $notificationJsonFromGrafana | jq --arg keys "$fieldsToKeep" 'with_entries( select( .key as $k | any($keys | fromjson[]; . == $k) ) )' > a.json

    equal=$(jq --argfile a a.json --argfile b b.json -n '$a == $b')

    # dunk some debug output to stdout if this is about to fail
    if [ "$equal" != "true" ]; then
        run diff <(jq -S . a.json) <(jq -S . b.json)
        echo $output
    fi

    [ "$equal" = "true" ]

    rm a.json
    rm b.json
}

# validatePostDataSource <yaml file name>
#   note that the notification file name must match the DataSource object name ...
#    ... b/c i'm lazy
#
validatePostDataSource() {
    specfile=$1

    sourceName=$(objectNameFromFile $specfile)

    # create in kubernetes
    kubectl apply -f $specfile >&2

	sleep 5s

    sourceId=$(kubectl get DataSource -o=jsonpath="{.items[?(@.metadata.name==\"${sourceName}\")].status.grafanaID}")

    [ "$sourceId" != "" ]

    # check if exists in grafana
	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/datasources/${sourceId})
    [ "$httpStatus" -eq "200" ]

    echo $sourceId
}

#
# validateDataSourceCount <count>
#   use grafana search api to confirm that the count is what is expected 
#
validateDataSourceCount() {
    sourceJson=$(curl --silent ${GRAFANA_URL}/api/datasources)

    count=$(echo $sourceJson | jq length)

    [ "$count" -eq "$1" ]
}

#
# validateDataSourceContents <yaml file name>
#   creates a DataSource and both verifies it exists and that its content matches
#
validateDataSourceContents() {
    filename=$1

    sourceId=$(validatePostDataSource $filename)

    echo "Test Json Content of $filename ($sourceId)"

    sourceJsonFromYaml=$(grep -A9999 'json' $filename)
    sourceJsonFromYaml=${sourceJsonFromYaml%?}   # strip final quote
    sourceJsonFromYaml=${sourceJsonFromYaml#*\'} # strip up to and including the first quote

    # remove the version field from comparison b/c grafana will update it
    echo $sourceJsonFromYaml | jq '. | del(.version)'  > b.json
    fieldsToKeep=$(cat b.json | jq keys)

    sourceJsonFromGrafana=$(curl --silent ${GRAFANA_URL}/api/datasources/${sourceId})

    # grafana can add fields to flesh out the object.  remove anything from grafana not present in the original
    #  spec file
    echo $sourceJsonFromGrafana | jq --arg keys "$fieldsToKeep" 'with_entries( select( .key as $k | any($keys | fromjson[]; . == $k) ) )' > a.json

    equal=$(jq --argfile a a.json --argfile b b.json -n '$a == $b')

    # dunk some debug output to stdout if this is about to fail
    if [ "$equal" != "true" ]; then
        run diff <(jq -S . a.json) <(jq -S . b.json)
        echo $output
    fi

    [ "$equal" = "true" ]

    rm a.json
    rm b.json
}

#
# validateMetrics <metric name> <object type> <value>
#   confirms the metric name exists for the object type and value...kind of
#
validateMetrics() {
    metrics=$(curl --silent ${CONTROLLER_URL}/metrics)

    if [ -n "$3" ]; then
        echo $metrics | grep $1 | grep $2 | grep $3
    else
        echo $metrics | grep $1 | grep $2
    fi
}

#
# utils
#
dumpState() {
    echo "-----------events--------------"
    kubectl get events -o=custom-columns=NAME:.involvedObject.name,KIND:.involvedObject.kind,REASON:.reason

    echo "-----------controller logs--------------"
    kubectl logs $(kubectl get po -l=run=kubernetes-grafana-test --no-headers=true | cut -d ' ' -f 1)
    echo "-----------grafana logs--------------"
    kubectl logs $(kubectl get po -l=app=grafana --no-headers=true| cut -d ' ' -f 1)

    echo "-----------Dashboards--------------"
    kubectl describe Dashboard
    echo "-----------AlertNotification--------------"
    kubectl describe AlertNotification
    echo "-----------DataSources --------------"
    kubectl describe DataSource
}

#
# objectNameFromFile <filename>
#
objectNameFromFile() {
    file=$1

    name="${file##*/}"
    name="${name%.*}"

    echo $name
}

#
# validateEvents <datatype> <eventname> <object name>
#
validateEvents() {
    kubectl get events -o=custom-columns=NAME:.involvedObject.name,KIND:.involvedObject.kind,REASON:.reason | grep $1 | grep $2 | grep $3
}
