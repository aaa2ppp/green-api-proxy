const PROXY_URL = "/green-api/proxy";

async function callApi(method) {
    const idInstance = document.getElementById("idInstance").value;
    const apiToken = document.getElementById("apiToken").value;

    let url, payload;

    switch (method) {
        case "getSettings":
            url = `/waInstance${idInstance}/getSettings`;
            break;

        case "getStateInstance":
            url = `/waInstance${idInstance}/getStateInstance`;
            break;

        case "sendMessage":
            const chatId = document.getElementById("chatId").value;
            const message = document.getElementById("message").value;
            url = `/waInstance${idInstance}/sendMessage`;
            payload = { chatId: chatId, message: message };
            break;

        case "sendFileByUrl":
            const fileChatId = document.getElementById("fileChatId").value;
            const fileUrl = document.getElementById("fileUrl").value;
            const fileName = document.getElementById("fileName").value;
            url = `/waInstance${idInstance}/sendFileByUrl`;
            payload = { chatId: fileChatId, urlFile: fileUrl, fileName: fileName };
            break;

        default:
            alert("Неизвестный метод");
            return;
    }

    document.getElementById("apiResponse").textContent = "";

    try {
        const options = {
            method: payload ? "POST" : "GET",
            headers: { "Content-Type": "application/json", "X-ApiToken": apiToken },
            body: payload ? JSON.stringify(payload) : undefined,
        };

        const proxiedUrl = PROXY_URL + url;
        const response = await fetch(proxiedUrl, options);
        if (!response.ok) {
            document.getElementById("apiResponse").textContent = `Ошибка: HTTP ${response.status}\n`;
        }

        const rawText = await response.text();

        try {
            const data = JSON.parse(rawText);
            maskApiToken(data);
            document.getElementById("apiResponse").textContent += JSON.stringify(data, null, 2);
        } catch (e) {
            document.getElementById("apiResponse").textContent += rawText;
        }
    } catch (error) {
        document.getElementById("apiResponse").textContent = `Ошибка: ${error.message}`;
    }
}

function getFilenameFromUrl(url) {
    const pathname = new URL(url).pathname;
    return pathname.split("/").pop();
}

function updateFileName() {
    url = document.getElementById("fileUrl").value;
    if (!url) {
        document.getElementById("fileName").value = "";
        return;
    }
    const fileName = getFilenameFromUrl(url) || "some_file";
    document.getElementById("fileName").value = fileName;
}

function maskApiToken(data) {
    if (!data.path) {
        return;
    }
    data.path = data.path.replace(/[^\/]*$/, "***********");
}
