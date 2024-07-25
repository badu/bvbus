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

    const loadBusTimetables = async (busId, okHandler, errorHandler) => {
        await fetch(`./rtt/${busId}.json`).then((response) => {
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
        loadBusTimetables,
    }
}