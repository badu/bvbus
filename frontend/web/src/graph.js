import distances from "@/distances.js"
import urban_busses from "@/urban_busses.js";

class Queue {
    #size

    constructor() {
        this.head = null
        this.tail = null
        this.#size = 0

        return Object.seal(this)
    }

    get length() {
        return this.#size
    }

    enqueue(data) {
        const node = {data, next: null}

        if (!this.head && !this.tail) {
            this.head = node
            this.tail = node
        } else {
            this.tail.next = node
            this.tail = node
        }

        return ++this.#size
    }

    dequeue() {
        if (this.isEmpty()) {
            throw new Error('Queue is Empty')
        }

        const firstData = this.peekFirst()

        this.head = this.head.next

        if (!this.head) {
            this.tail = null
        }

        this.#size--

        return firstData
    }

    peekFirst() {
        if (this.isEmpty()) {
            throw new Error('Queue is Empty')
        }

        return this.head.data
    }

    isEmpty() {
        return this.length === 0
    }
}

class Graph {
    constructor() {
        this.edges = []
        this.nodes = new Set()
        this.neighbors = new Map()
    }

    addEdge(from, to, weight) {
        if (!this.nodes.has(from)) {
            this.nodes.add(from)
        }

        if (!this.nodes.has(to)) {
            this.nodes.add(to)
        }

        if (!this.neighbors.has(from)) {
            this.neighbors.set(from, new Set())
        }

        this.neighbors.get(from).add(to)

        this.edges.push({from: from, to: to, weight: weight})
    }

    findRoute(from, to) {
        // check if startNode & targetNode are identical
        if (from === to) {
            return [from]
        }

        // visited keeps track of all nodes visited
        const visited = new Set()

        // queue contains the paths to be explored in the future
        const initialPath = [from]
        const queue = new Queue()
        queue.enqueue(initialPath)

        while (!queue.isEmpty()) {
            // start with the queue's first path
            const path = queue.dequeue()
            const node = path[path.length - 1]

            // explore this node if it hasn't been visited yet
            if (!visited.has(node)) {
                // mark the node as visited
                visited.add(node)

                if (!this.neighbors.has(node)) {
                    console.log(`${node} doesn't have any neighbours. skipping...`)
                    continue
                }

                const neighbors = this.neighbors.get(node)
                // create a new path in the queue for each neighbor
                for (const node of neighbors) {
                    const newPath = path.concat([node])

                    // the first path to contain the target node is the shortest path
                    if (node === to) {
                        return newPath
                    }

                    // queue the new path
                    queue.enqueue(newPath)
                }
            }
        }

        // the target node was not reachable
        return []
    }
}

const stationsAndBusses = new Map()
const graph = new Graph()
for (let sIdx = 0; sIdx < distances.length; sIdx++) {
    for (let tIdx = 0; tIdx < distances[sIdx].s.length; tIdx++) {
        graph.addEdge(distances[sIdx].i, distances[sIdx].s[tIdx].t, distances[sIdx].s[tIdx].m)
        for (let j = 0; j < urban_busses.length; j++) {
            for (let k = 1; k < urban_busses[j].s.length - 1; k++) {
                if (urban_busses[j].s[k - 1] === distances[sIdx].i && urban_busses[j].s[k] === distances[sIdx].s[tIdx].t) {
                    const stationKey = `${distances[sIdx].i}-${distances[sIdx].s[tIdx].t}`
                    if (!stationsAndBusses.has(stationKey)) {
                        stationsAndBusses.set(stationKey, [])
                    }
                    stationsAndBusses.get(stationKey).push(urban_busses[j].i)
                }
            }
        }
    }
}

export const pathFinding = () => {
    return {
        graph, stationsAndBusses
    }
}