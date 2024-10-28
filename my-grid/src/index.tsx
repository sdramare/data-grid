// ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
// ┃ ██████ ██████ ██████       █      █      █      █      █ █▄  ▀███ █       ┃
// ┃ ▄▄▄▄▄█ █▄▄▄▄▄ ▄▄▄▄▄█  ▀▀▀▀▀█▀▀▀▀▀ █ ▀▀▀▀▀█ ████████▌▐███ ███▄  ▀█ █ ▀▀▀▀▀ ┃
// ┃ █▀▀▀▀▀ █▀▀▀▀▀ █▀██▀▀ ▄▄▄▄▄ █ ▄▄▄▄▄█ ▄▄▄▄▄█ ████████▌▐███ █████▄   █ ▄▄▄▄▄ ┃
// ┃ █      ██████ █  ▀█▄       █ ██████      █      ███▌▐███ ███████▄ █       ┃
// ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
// ┃ Copyright (c) 2017, the Perspective Authors.                              ┃
// ┃ ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌ ┃
// ┃ This file is part of the Perspective library, distributed under the terms ┃
// ┃ of the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0). ┃
// ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

import * as React from "react";
import * as ReactDOM from "react-dom";
import * as perspective from "@finos/perspective";

import "@finos/perspective-viewer";
import "@finos/perspective-viewer-datagrid";
import "@finos/perspective-viewer-d3fc";
import {
    HTMLPerspectiveViewerElement,
} from "@finos/perspective-viewer";

import "./index.css";

const worker = await perspective.default.worker();

const getTable = async (): Promise<perspective.Table> => {
    const req = fetch("http://localhost:8081/table");
    const resp = await req;
    const buffer = await resp.arrayBuffer();
    return await worker.table(buffer, { index: "ID" });
};

const config = {
    "title": "DWE Report",
    "columns_config": {
        "dwe %1": {
            "number_bg_mode": "pulse",
            "number_format": {
                "minimumFractionDigits": 2,
                "maximumFractionDigits": 5
            }
        }
    },

};

const App = (): React.ReactElement => {
    const viewer = React.useRef<HTMLPerspectiveViewerElement>(null);

    React.useEffect(() => {
        getTable().then(async table => {

            setInterval(async () => {
                const update = await fetch("http://localhost:8081/update")
                const buffer = await update.arrayBuffer()
                table.update(buffer)
            }, 1000)

            if (viewer.current) {
                viewer.current.load(table);
                viewer.current.restore(config);
            }
        });
    }, []);

    return <perspective-viewer ref={viewer}></perspective-viewer>;
};

console.log("render")
ReactDOM.render(<App />, document.getElementById("root"));