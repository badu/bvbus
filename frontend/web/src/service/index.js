export const service = (toast) => {
    const loadStationTimetables = async (stationId, targetStation, okHandler, errorHandler) => {
        await fetch(`./tt/${stationId}.json`).then((response) => {
            const contentType = response.headers.get("content-type")
            if (response.ok) {
                if (contentType && contentType.indexOf("application/json") !== -1) {
                    try {
                        return response.json()
                    } catch (e) {
                        console.error("json parse error", e)
                    }
                    return null
                } else {
                    console.error(`response is not json loading timetables for ${stationId}`)
                    return null
                }
            } else {
                console.error("response", response)
                return null
            }
        }).then((data) => {
            if (data) {
                okHandler(data, targetStation)
            } else {
                errorHandler()
            }
        })
    }

    const loadStreetPoints = async (stationsPairIds, okHandler, errorHandler) => {
        console.log(`loading ${stationsPairIds}`)
        await fetch(`./pt/${stationsPairIds}.json`).then((response) => {
            const contentType = response.headers.get("content-type")
            if (response.ok) {
                if (contentType && contentType.indexOf("application/json") !== -1) {
                    try {
                        return response.json()
                    } catch (e) {
                        console.error("json parse error", e)
                    }
                    return null
                } else {
                    console.error(`response is not json, loading street points ${stationsPairIds}`)
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

    const loadDirectPathFinder = async (stationId, okHandler, errorHandler) => {
        await fetch(`./pf/${stationId}.json`).then((response) => {

            const contentType = response.headers.get("content-type")
            if (response.ok) {
                if (contentType && contentType.indexOf("application/json") !== -1) {
                    try {
                        return response.json()
                    } catch (e) {
                        console.error("json parse error", e)
                    }
                    return null
                } else {
                    console.error(`response is not json loading path finders for ${stationId}`)
                    return null
                }
            } else {
                return null
            }
        }).then((data) => {
            if (data) {
                okHandler(data)
            } else {
                console.error('error')
                errorHandler()
            }
        })
    }

    return {
        loadStationTimetables,
        loadStreetPoints,
        loadDirectPathFinder,
    }
}