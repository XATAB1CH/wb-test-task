function showAlert(message, type = "danger") {
    const container = document.getElementById("alertContainer");
    const alert = document.createElement("div");
    alert.className = `alert alert-${type} alert-dismissible fade show`;
    alert.role = "alert";
    alert.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Закрыть"></button>
    `;
    container.appendChild(alert);

    setTimeout(() => {
        alert.classList.remove("show");
        alert.classList.add("hide");
        alert.addEventListener("transitionend", () => alert.remove());
    }, 3000);
}

function syntaxHighlight(json) {
    if (typeof json != 'string') {
        json = JSON.stringify(json, undefined, 2);
    }
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|\b\d+\.?\d*\b)/g, function (match) {
        let cls = 'json-number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'json-key';
            } else {
                cls = 'json-string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'json-boolean';
        } else if (/null/.test(match)) {
            cls = 'json-null';
        }
        return `<span class="${cls}">${match}</span>`;
    });
}

document.getElementById('orderForm').addEventListener('submit', async function(e) {
    e.preventDefault();
    const orderId = document.getElementById('orderId').value;
    const orderResult = document.getElementById('orderResult');

    orderResult.style.display = "none";
    orderResult.innerHTML = "";

    try {
        const response = await fetch(`http://localhost:8081/order/${orderId}`);
        if (!response.ok) {
            if (response.status === 404) {
                showAlert("Заказ с таким UID не найден", "warning");
            } else {
                showAlert(`⚠ Ошибка сервера: ${response.status}`, "danger");
            }
            return;
        }

        const data = await response.json();
        orderResult.innerHTML = `<pre><code>${syntaxHighlight(data)}</code></pre>`;
        orderResult.style.display = "block";
    } catch (err) {
        showAlert("⚠ Ошибка соединения с сервером", "danger");
    }
});
