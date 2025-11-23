
var token = localStorage.getItem('token');

async function getDictionary(url) {
    return fetch(url, {
        method: 'GET',
    })
        .then(response => response.json())
        .then(data => {
            console.log(url);
            console.log(data);
            populateDictionary(data);
        })
        .catch(error => {
            console.error('Ошибка', error);
        });
}

async function get(url, token) {
    return fetch(url, {
        method: 'GET',
        headers: new Headers({
            "Authorization": `Bearer ${token}`
        }),
    })
        .then(response => response.json())
        .then(data => {
            console.log(url);
            console.log(data);
            showTable(true);
            const tableBody = document.getElementById("tableBody");
            tableBody.innerHTML = "";
            const tableFooter = document.getElementById("tableFooter");
            tableFooter.innerHTML = "";
            populateTable(data.records);
            populateFooter(data.summaryByRoot);
            const errorMessage = document.getElementById('errorMessage');
            errorMessage.textContent = '';
        })
        .catch(error => {
            showTable(false);
            console.error('Ошибка', error);
            const errorMessage = document.getElementById('errorMessage');
            errorMessage.textContent = 'Произошла ошибка при создании отчёта.';
        });
}

async function populateDictionary(dictionaries) {
    const selectDictionary = document.getElementById('select-1');
    dictionaries.forEach(dictionary => {
        var option = document.createElement('option');
        option.value = dictionary.id;
        option.text = dictionary.code;
        selectDictionary.appendChild(option);
        console.log(option.text);
    });
}

const urlDictionary = `https://mis-api.kreosoft.space/api/dictionary/icd10/roots`;
getDictionary(urlDictionary);


document.getElementById("submitButton").addEventListener("click", function () {
    let icdRoot = '';
    const inputDataStart = document.getElementById("inputDataStart").value;
    const inputDataEnd = document.getElementById("inputDataEnd").value;
    const selectOptions = document.getElementById("select-1").selectedOptions;
    let icdRoots = "";
    for (let i = 0; i < selectOptions.length; i++) {
        icdRoots += selectOptions[i].value;
        if (i !== selectOptions.length - 1) {
            icdRoots += "&icdRoots=";
        }
        if (!selectOptions.length == 0) {
            icdRoot = '&icdRoots=';
        }
    }
    const url = `https://mis-api.kreosoft.space/api/report/icdrootsreport?start=${inputDataStart}&end=${inputDataEnd}${icdRoot}${icdRoots}`;
    get(url, token);
});

function populateTable(data) {
    data.forEach(record => {
        const row = document.createElement("tr");
        const patientNameCell = document.createElement("td");
        patientNameCell.textContent = record.patientName;
        const patientBirthdateCell = document.createElement("td");
        patientBirthdateCell.textContent = formatDataTime(record.patientBirthdate);
        const genderCell = document.createElement("td");
        genderCell.textContent = record.gender === 'Male' ? 'Пол: Мужской' : record.gender === 'Female' ? 'Пол: Женский' : 'Не указан';
        const visitsByRootCell = document.createElement("td");
        const visitsByRoot = record.visitsByRoot;
        Object.entries(visitsByRoot).forEach(([key, value]) => {
            const visit = document.createElement("p");
            visit.textContent = `${key}: ${value}`;
            visit.classList.add("background2");
            visitsByRootCell.appendChild(visit);
        });
        row.appendChild(patientNameCell);
        row.appendChild(patientBirthdateCell);
        row.appendChild(genderCell);
        row.appendChild(visitsByRootCell);
        tableBody.appendChild(row);
    });
}

function populateFooter(summaryByRoot) {
    const row = document.createElement("tr");
    const totalCell = document.createElement("td");
    totalCell.textContent = "Общее число посещений";
    const visitsByRootCell = document.createElement("td");
    const visitsByRoot = Object.entries(summaryByRoot);
    visitsByRootCell.setAttribute("colspan", "3");
    visitsByRoot.forEach(([key, value]) => {
        const visit = document.createElement("p");
        visit.textContent = `${key}: ${value}`;
        visit.classList.add("background2");
        visitsByRootCell.appendChild(visit);
    });
    row.appendChild(totalCell);
    row.appendChild(visitsByRootCell);
    tableFooter.appendChild(row);
}


function formatDataTime(originalDate) {
    const dateParts = originalDate.split("T")[0].split("-");
    const timeParts = originalDate.split("T")[1].split(":");
    const day = dateParts[2];
    const month = dateParts[1];
    const year = dateParts[0];
    const hours = timeParts[0];
    const minutes = timeParts[1];
    const formattedDate = `${day}.${month}.${year} ${hours}:${minutes}`;
    return formattedDate;
}

function showTable(bool) {
    const tableContainer = document.getElementById("tableContainer");
    const dataTable = document.getElementById("dataTable");
    if (bool) {
        tableContainer.style.display = "block";
        dataTable.style.display = "table";
    } else {
        tableContainer.style.display = "none";
        dataTable.style.display = "none";
    }
}


//--------------------------Multiselect-------------------------------------

let multiselect_block = document.querySelectorAll(".multiselect_block");
multiselect_block.forEach(parent => {
    let label = parent.querySelector(".field_multiselect");
    let select = parent.querySelector(".field_select");
    let text = label.innerHTML;
    select.addEventListener("change", function (element) {
        let selectedOptions = this.selectedOptions;
        label.innerHTML = "";
        for (let option of selectedOptions) {
            let button = document.createElement("button");
            button.type = "button";
            button.className = "btn_multiselect";
            button.textContent = option.text;
            button.onclick = _ => {
                option.selected = false;
                button.remove();
                if (!select.selectedOptions.length) label.innerHTML = text
            };
            label.append(button);
        }
    })
})
