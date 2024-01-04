// I wish I added jQuery to this project ):

var ws;

function get_code() {
    return window.location.pathname.split('/')[2];
}

function getCookie(cname) {
    let name = cname + "=";
    let decodedCookie = decodeURIComponent(document.cookie);
    let ca = decodedCookie.split(';');
    for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
            return c.substring(name.length, c.length);
        }
    }
    return "";
}

// 'num_subnets': int
// 'subnets': map[int][int]string
// 'ip_addresses': map[string]string
function handle_metadata(metadata) {
    let num_subnets = metadata.num_subnets;
    let subnets = metadata.subnets;
    let ip_addresses = metadata.ip_addresses;

    // create a button to join each subnet
    let join_subnet = function (subnet_id) {
        return function () {
            send_message({
                type: "JoinSubnet",
                payload: {
                    subnet: parseInt(subnet_id),
                },
            });
        };
    }

    // create a button for each subnet
    let subnet_table = document.getElementById("subnet-table");
    subnet_table.innerHTML = "";

    // create a row for the join buttons
    let join_row = document.createElement("tr");
    subnet_table.appendChild(join_row);

    // create a join button for each subnet
    for (let subnet_id = 1; subnet_id <= num_subnets; subnet_id++) {
        let join_button = document.createElement("button");
        join_button.innerHTML = "Join 192.168." + subnet_id + ".0/24";
        join_button.onclick = join_subnet(subnet_id);
        let join_cell = document.createElement("td");
        join_cell.appendChild(join_button);
        join_row.appendChild(join_cell);
    }

    // bring subnet maps into lists of names + ips
    let subnets_lists = [];
    for (let subnet_id = 1; subnet_id <= num_subnets; subnet_id++) {
        let subnet_list = [];
        for (let ip = 1; ip <= 255; ip++) {
            let name = subnets[subnet_id][ip];
            if (name != undefined) {
                subnet_list.push({
                    "name": name,
                    "ip": ip_addresses[name],
                });
            }
        }
        subnets_lists.push(subnet_list);
    }

    // find the max number of rows
    let rows = subnets_lists.map(function (subnet) {
        return subnet.length;
    }).reduce(function (a, b) {
        return Math.max(a, b);
    });

    // creates rows
    for (let row = 0; row < rows; row++) {
        let subnet_row = document.createElement("tr");
        subnet_table.appendChild(subnet_row);

        // create a cell for each subnet
        for (let subnet_id = 1; subnet_id <= num_subnets; subnet_id++) {
            let subnet = subnets_lists[subnet_id - 1];
            let subnet_cell = document.createElement("td");
            subnet_row.appendChild(subnet_cell);

            // if there is a name in this subnet, add it
            if (subnet[row] != undefined) {
                // trim the leading and trailing quotes
                let name = subnet[row].name.substring(1, subnet[row].name.length - 1);

                let ip = subnet[row].ip;
                let text = name + " (" + ip + ")";
                let text_node = document.createTextNode(text);
                let text_span = document.createElement("span");
                text_span.appendChild(text_node);
                subnet_cell.appendChild(text_span);
            }
        }
    }
}

// Name Name `json:"name"`
// IP IP `json:"ip,omitempty"`
// Score int `json:"score,omitempty"`
// QATable QATable `json:"qa_table"`
// Challenges []Challenge `json:"challenges,omitempty"`

function handle_userdata(userdata) {
    let name = userdata.name.substring(1, userdata.name.length - 1);

    // whois docuement
    let whois = document.getElementById("whois");
    whois.innerHTML = "";

    // name + ip (if available)
    var name_node;
    if (userdata.ip != undefined) {
        let ip = userdata.ip;
        let text = name + " (" + ip + ")";
        name_node = document.createTextNode(text);
    } else {
        name_node = document.createTextNode(name);
    }
    let name_span = document.createElement("span");
    name_span.appendChild(name_node);
    whois.appendChild(name_span);

    // fill out qa-table
    let qa_table = document.getElementById("qa-table");
    qa_table.innerHTML = "";

    // header row
    let header_row = document.createElement("tr");
    let question_header = document.createElement("th");
    question_header.innerHTML = "Questions";
    let answer_header = document.createElement("th");
    answer_header.innerHTML = "Answers";
    header_row.appendChild(question_header);
    header_row.appendChild(answer_header);
    qa_table.appendChild(header_row);

    for (let question in userdata.qa_table) {
        let answer = userdata.qa_table[question];

        let question_node = document.createTextNode(question);
        let question_span = document.createElement("td");
        question_span.appendChild(question_node);

        let answer_node = document.createTextNode(answer);
        let answer_span = document.createElement("td");
        answer_span.appendChild(answer_node);

        let row = document.createElement("tr");
        row.appendChild(question_span);
        row.appendChild(answer_span);
        qa_table.appendChild(row);
    }

    // If there are challenges, show them
}

async function ws_connect() {
    // check if we a session cookie
    let session = getCookie("session");
    console.log("Session cookie: " + session);

    if (session == "") {
        let register_path = "/room/" + get_code() + "/register";
        console.log("Registering user with path " + register_path);
        await fetch(register_path);
    }

    // ws or wss
    let ws_scheme = window.location.protocol === "https:" ? "wss" : "ws";
    let ws_path = ws_scheme + "://" + window.location.host + "/room/" + get_code() + "/ws";
    ws = new WebSocket(ws_path);

    // on open handler
    ws.onopen = function () {
        console.log("Connected to ws");
        send_message({
            type: "RequestMetadata",
            payload: {},
        });

        send_message({
            type: "WhoAmI",
            payload: {},
        });
    };

    // on message handler
    ws.onmessage = function (event) {
        let data = JSON.parse(event.data);
        console.log(data);

        // switch on message type
        switch (data.type) {
            case "AssignedIP":
                break;
            case "CreateChallenge":
                break;
            case "Grade":
                break;
            case "Metadata":
                handle_metadata(data.payload);
                break;
            case "Userdata":
                handle_userdata(data.payload);
                break;
        }
    };

    // on close handler
    ws.onclose = function () {
        console.log("Disconnected from ws");
        // try to reconnect in 1 seconds
        setTimeout(ws_connect, 5000);
    };
}

function send_message(msg) {
    ws.send(JSON.stringify(msg));
}

// on load handler
window.onload = function () {
    ws_connect();
};