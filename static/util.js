export{GenerateUniqueID}
function GenerateUniqueID(){
    return (new Date()).valueOf().toString(16) + Math.random(10).toString(16).substring(2)
}