function(task, responses){
    if(responses.length === 0){
        return {"plaintext": "No response yet from agent..."};
    }
    if(task.status.includes("error")){
        return {"plaintext": responses.join("\n")};
    }
    try{
        let collection = JSON.parse(responses[0]);
        let headers = [
            {"plaintext": "DL", "type": "button", "width": 50, "disableSort": true},
            {"plaintext": "Reg", "type": "button", "width": 50, "disableSort": true},
            {"plaintext": "Name", "type": "string", "fillWidth": true},
            {"plaintext": "Command", "type": "string", "fillWidth": true},
            {"plaintext": "Description", "type": "string", "fillWidth": true}
        ];
        let rows = [];
        for(let i = 0; i < collection.length; i++){
            rows.push({
                "DL": {"button":{
                        "name": "",
                        "type": "task",
                        "ui_feature": "forge:download",
                        "hoverText": collection[i]["downloadable"] ? collection[i]["downloaded"] ? "Re-download latest from GitHub and register command" : "Download latest from GitHub and register command" : "No url associated with this source, so it can't be re downloaded",
                        "disabled": !collection[i]["downloadable"],
                        "startIcon": collection[i]["downloadable"] ? collection[i]["downloaded"] ? "refresh" : "download" : collection[i]["downloaded"] ? "check" : "x" ,
                        "startIconColor": collection[i]["downloadable"] ? collection[i]["downloaded"] ? "info" : "warning" : collection[i]["downloaded"] ? "success" : "error" ,
                        "parameters": {"collectionName": collection[i]["collection_name"], "commandName": collection[i]["name"]}
                    }},
                "Reg": {"button": {
                        "name": "",
                        "type": "task",
                        "ui_feature": "forge:register",
                        "hoverText": collection[i]["registered"] ? "Remove Command": "Register command",
                        "startIcon": collection[i]["registered"] ? "x": "add",
                        "startIconColor": collection[i]["registered"] ? "error": "success",
                        "parameters": {"collectionName": collection[i]["collection_name"], "commandName": collection[i]["name"], "remove": collection[i]["registered"]}
                    }},
                "Name": {"plaintext": collection[i]["name"]},
                "Command": {"plaintext": collection[i]["command_name"]},
                "Description": {"plaintext": collection[i]["description"]}
            });
        }
        return {"table": [{
                "headers": headers,
                "rows": rows
            }]}
    }catch(error){
        console.log(error)
        return {"plaintext": responses.join("\n")};
    }
}