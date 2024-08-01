export const service = () => {
    const loadStationTimetables = async (stationId, okHandler, errorHandler) => {
        await fetch(`./tt/${stationId}.json`).then((response) => {
            const contentType = response.headers.get("content-type")
            if (response.ok) {
                if (contentType && contentType.indexOf("application/json") !== -1) {
                    return response.json()
                } else{
                    return null
                }
            } else {
                return null
            }
        }).then((data) => {
            if (data) {
                okHandler(data)
            } else {
                errorHandler()
            }
        })
    }

    const loadBusPoints = async (busId, okHandler, errorHandler) => {
        await fetch(`./pt/${busId}.json`).then((response) => {
            const contentType = response.headers.get("content-type")
            if (response.ok) {
                if (contentType && contentType.indexOf("application/json") !== -1) {
                    return response.json()
                } else{
                    return null
                }
            } else {
                return null
            }
        }).then((data) => {
            if (data) {
                okHandler(data)
            } else {
                errorHandler()
            }
        })
    }

    const loadDirectPathFinder = async(stationId, okHandler, errorHandler)=>{
        await fetch(`./pf/${stationId}.json`).then((response) => {
            const contentType = response.headers.get("content-type")
            if (response.ok) {
                if (contentType && contentType.indexOf("application/json") !== -1) {
                    return response.json()
                } else{
                    return null
                }
            } else {
                return null
            }
        }).then((data) => {
            if (data) {
                okHandler(data)
            } else {
                errorHandler()
            }
        })
    }

    const loadIndirectPathFinder = async(stationId, okHandler, errorHandler)=>{
        await fetch(`./pf/${stationId}-cross.json`).then((response) => {
            const contentType = response.headers.get("content-type")
            if (response.ok) {
                if (contentType && contentType.indexOf("application/json") !== -1) {
                    return response.json()
                } else{
                    return null
                }
            } else {
                return null
            }
        }).then((data) => {
            if (data) {
                okHandler(data)
            } else {
                errorHandler()
            }
        })
    }

    return {
        loadStationTimetables,
        loadBusPoints,
        loadDirectPathFinder,
        loadIndirectPathFinder,
    }
}