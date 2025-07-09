
const PROXY_URL = "/green-api-proxy";

async function callApi(method) {
    const idInstance = document.getElementById("idInstance").value;
    const apiToken = document.getElementById("apiToken").value;

    let url, payload;

    switch (method) {
        case "getSettings":
            url = `/waInstance${idInstance}/getSettings/${apiToken}`;
            break;

        case "getStateInstance":
            url = `/waInstance${idInstance}/getStateInstance/${apiToken}`;
            break;

        case "sendMessage":
            const chatId = document.getElementById("chatId").value;
            const message = document.getElementById("message").value;
            url = `/waInstance${idInstance}/sendMessage/${apiToken}`;
            payload = { chatId: chatId, message: message };
            break;

        case "sendFileByUrl":
            const fileChatId = document.getElementById("fileChatId").value;
            const fileUrl = document.getElementById("fileUrl").value;
            const fileName = getFilenameFromUrl(fileUrl) || "some_file";
            url = `/waInstance${idInstance}/sendFileByUrl/${apiToken}`;
            payload = { chatId: fileChatId, urlFile: fileUrl, fileName: fileName };
            break;

        default:
            alert("Неизвестный метод");
            return;
    }

    try {
        const options = {
            method: payload ? "POST" : "GET",
            headers: { "Content-Type": "application/json" },
            body: payload ? JSON.stringify(payload) : undefined
        };

        const proxiedUrl = PROXY_URL + url;
        const response = await fetch(proxiedUrl, options);

        if (!response.ok) {
            document.getElementById("apiResponse").textContent = `Ошибка: HTTP ${response.status}`;
            const rawText = await response.text();
            document.getElementById("apiResponse").textContent += rawText;
        } else {
            const data = await response.json();
            document.getElementById("apiResponse").textContent = JSON.stringify(data, null, 2);
        }
    } catch (error) {
        document.getElementById("apiResponse").textContent = `Ошибка: ${error.message}`;
    }
}

function getFilenameFromUrl(url) {
    const pathname = new URL(url).pathname;
    return pathname.split('/').pop();
}
