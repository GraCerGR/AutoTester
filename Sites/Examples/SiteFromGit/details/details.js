var token = localStorage.getItem('token');

let diagnosCount = 1;
let numberOfObjects2 = 0;

const url = new URL(window.location.href);
console.log(url);
const params = new URLSearchParams(url.search);

const inspectionId = params.get('inspectionId');
const patient = params.get('patient');

const urlPatient = `https://mis-api.kreosoft.space/api/inspection/${inspectionId}`;
console.log(urlPatient);
getInspecrionInfo(urlPatient, token);

async function getInspecrionInfo(url, token) {
    return fetch(url, {
        method: 'GET',
        headers: new Headers({
            "Authorization": `Bearer ${token}`
        }),
    })
        .then(response => response.json())
        .then(async data => {
            console.log(data)
            document.querySelector('[data-data]').textContent = `Амбулаторный осмотр от ` + await formatDataTime(data.date);
            document.querySelector('[data-name]').textContent = 'Пациент: ' + data.patient.name;
            document.querySelector('[data-gender]').textContent = data.patient.gender === 'Male' ? 'Пол: Мужской' : data.patient.gender === 'Female' ? 'Пол: Женский' : '';
            document.querySelector('[data-birthday]').textContent = `Дата рождения: ${data.patient.birthday ? await formatBirthday(data.patient.birthday.split('T')[0]) : 'Не указано'}`;
            document.querySelector('[data-doctor]').textContent = `Медицинский работник: ` + data.doctor.name;

            document.querySelector('[data-complaints]').textContent = data.complaints;
            document.querySelector('[data-anamnesis]').textContent = data.anamnesis;



            document.querySelector('[data-treatment]').textContent = data.treatment;
            //document.querySelector('[data-conclusion]').textContent = data.conclusion === 'Disease' ? 'Болезнь' : data.conclusion === 'Recovery' ? 'Выздоровление' : data.conclusion === 'Death' ? 'Смерть' : '';
            var conclusionElement = document.querySelector('[data-conclusion]');
            var conclusionText = '';

            if (data.conclusion === 'Disease') {
                conclusionText = 'Болезнь';
                document.querySelector('[data-visit]').textContent = 'Дата следующего визита: ' + await formatDataTime(data.nextVisitDate);
            } else if (data.conclusion === 'Death') {
                conclusionText = 'Смерть';
                document.querySelector('[data-visit]').textContent = 'Дата и время смерти: ' + await formatDataTime(data.deathDate);
            } else if (data.conclusion === 'Recovery') {
                conclusionText = 'Выздоровление';
            }

            conclusionElement.innerHTML = conclusionText;

            consultations(data);
            diagnoses(data.diagnoses)

            //document.querySelector('[data-consultations]').textContent = `Специализация консультанта: ` + data.consultations[i].speciality.name;

        })
        .catch(error => {
            console.error('Ошибка', error);
        });
}

async function getIdFromICD10(request) {
    const url = `https://mis-api.kreosoft.space/api/dictionary/icd10?request=${request}&page=1&size=5`;

    try {
        const response = await fetch(url);
        const data = await response.json();
        const id = data.records[0].id;
        return id;
    } catch (error) {
        console.error('Error:', error);
        return null;
    }
}

async function getDiagnosis(url, i) {
    return fetch(url, {
        method: 'GET',
    })
        .then(response => response.json())
        .then(data => {
            console.log(url);
            console.log(data);
            const datalistElement = document.getElementById(`diagnoses${i}`);
            datalistElement.innerHTML = '';
            data.records.forEach((diagnosis) => {
                const optionElement = document.createElement('option');
                optionElement.value = diagnosis.code;
                optionElement.text = diagnosis.name;
                datalistElement.appendChild(optionElement);
            });
        })
        .catch(error => {
            console.error('Ошибка', error);
        });
}


async function formatBirthday(originalDate) {
    const dateParts = originalDate.split("-");
    console.log(originalDate)
    const day = dateParts[2];
    const month = dateParts[1];
    const year = dateParts[0];
    originalDate = `${day}.${month}.${year}`;
    console.log(originalDate);
    return originalDate
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

//--------------Отрисовка консультаций
function consultations(data) {
    var consultationsContainer = document.querySelector('[data-consultations]');

    for (var i = 0; i < data.consultations.length; i++) {
        var consultation = data.consultations[i];

        var container = document.createElement('div');
        container.classList.add('container', 'container2', 'py-2', 'my-2');

        var label = document.createElement('label');
        label.classList.add('form-label', 'py-1', 'fw-bold', 'blue');
        label.textContent = 'Консультация';

        var consultationElement = document.createElement('div');
        consultationElement.classList.add('col', 'd-flex', 'align-items-center');
        consultationElement.innerHTML = 'Специализация консультанта: ' + consultation.speciality.name;

        var doctorElement = document.createElement('div');
        doctorElement.classList.add('fw-light');
        doctorElement.setAttribute('data-doctor', '');

        container.appendChild(label);
        container.appendChild(consultationElement);
        container.appendChild(doctorElement);

        consultationsContainer.appendChild(container);
    }
}

//--------------Отрисовка диагнозов

function diagnoses(diagnosesData) {
    var diagnosesContainer = document.querySelector('[data-diagnoses]');
    var diagnosesHTML = '';

    for (var i = 0; i < diagnosesData.length; i++) {
        var diagnosis = diagnosesData[i];

        var html = `
            <div class="container container3 py-2 my-2">
                <div class="col d-flex align-items-center">
                    <h6>(${diagnosis.code}) ${diagnosis.name}</h6>
                </div>
                <div class="fw-light">Тип в осмотре: ${diagnosis.type === 'Main' ? 'Основной' : diagnosis.type === 'Concomitant' ? 'Сопутствующий' : 'Осложнение'}</div>
                <div class="fw-light">Расшифровка: ${diagnosis.description}</div>
            </div>
        `;

        diagnosesHTML += html;
    }

    diagnosesContainer.innerHTML = diagnosesHTML;
}

//--------------------------Заключения
function handleConclusionChange() {
    const conclusionSelect = document.getElementById('conclusionSelect');
    const diseaseDateField = document.getElementById('diseaseDateField');
    const deathDateField = document.getElementById('deathDateField');
    document.getElementById('diseaseDate').value = null;
    document.getElementById('deathDate').value = null;

    if (conclusionSelect.value === 'Disease') {
        diseaseDateField.style.display = 'block';
        deathDateField.style.display = 'none';
    } else if (conclusionSelect.value === 'Death') {
        diseaseDateField.style.display = 'none';
        deathDateField.style.display = 'block';
    } else {
        diseaseDateField.style.display = 'none';
        deathDateField.style.display = 'none';
    }
}

//----------------------Диагнозы

const inputElement1 = document.getElementById('selectDiagnoses1');
inputElement1.addEventListener('input', function () {
    const input = event.target.value;
    const url = `https://mis-api.kreosoft.space/api/dictionary/icd10?request=${input}&page=1&size=5`;
    getDiagnosis(url, 1);
});


document.getElementById('addDiagnosBtn').addEventListener('click', function () {
    diagnosCount++;
    const diagnosId = diagnosCount;
    const newDiagnos = document.createElement('div');
    newDiagnos.id = diagnosId;
    newDiagnos.innerHTML = `
        <div class="d-flex align-items-center">
            <div class="col">
                <label class="form-label fw-light py-1">Болезни</label>
                <input type="text" class="form-control" placeholder="" id="selectDiagnoses${diagnosCount}" list="diagnoses${diagnosCount}">
                <datalist id="diagnoses${diagnosCount}">
                </datalist>
            </div>
        </div>
        <div id="${diagnosId}DiagnosComment">
            <textarea class="form-control my-2" rows="1"></textarea>
        </div>
        <label class="form-label fw-light py-1">Тип диагноза в осмотре</label>
        <div class="form-group">
          <div class="form-check form-check-inline">
            <input class="form-check-input" type="radio" name="Radios${diagnosCount}" id="Main${diagnosCount}" value="Main">
            <label class="form-check-label">
              Основной
            </label>
          </div>
          <div class="form-check form-check-inline">
            <input class="form-check-input" type="radio" name="Radios${diagnosCount}" id="Concomitant${diagnosCount}" value="Concomitant" checked>
            <label class="form-check-label">
              Сопутствующий
            </label>
          </div>
          <div class="form-check form-check-inline">
            <input class="form-check-input" type="radio" name="Radios${diagnosCount}" id="Complication${diagnosCount}" value="Complication">
            <label class="form-check-label">
              Осложнение
            </label>
          </div>
        </div>
        <p id="diagnos1"></p>
        <div class="col my-2" id="addDiagnos">
            <button class="btn btn-danger delete-diagnos-btn my-1" data-diagnos-id="${diagnosId}">- Удалить диагноз</button>
        </div>
    `;
    const diagnosContainer = document.getElementById('diagnos1');
    diagnosContainer.parentNode.insertBefore(newDiagnos, diagnosContainer.nextSibling);

    const inputElement = document.getElementById(`selectDiagnoses${diagnosCount}`);
    inputElement.addEventListener('input', function () {
        const input = event.target.value;
        const url = `https://mis-api.kreosoft.space/api/dictionary/icd10?request=${input}&page=1&size=5`;
        getDiagnosis(url, diagnosCount);
    });
    ;
    numberOfObjects2 = document.querySelectorAll('.delete-diagnos-btn').length;
});

//---------------------------------Кнопки удаления диагноза и сохранения изменений

document.addEventListener('click', function (event) {
    if (event.target.classList.contains('delete-diagnos-btn')) {
        const diagnosId = event.target.dataset.diagnosId;
        const diagnosElement = document.getElementById(diagnosId);
        if (diagnosElement) {
            diagnosElement.remove();
        }
    } else if (event.target.classList.contains('class')) {
        InspectionChange();
        const registrationModal = new bootstrap.Modal(document.getElementById('registrationModal'));
        registrationModal.hide();
    }
});

//-----------------------------Изменение осмотра

async function InspectionChange() {
    let diagnoses = [];
    console.log(numberOfObjects2);
    for (let i = 1; i <= numberOfObjects2 + 1; i++) {
        let icdDiagnosisId = document.getElementById(`selectDiagnoses${i}`).value;
        let description = document.getElementById(`${i}DiagnosComment`).querySelector('textarea').value;
        let type = document.querySelector(`input[name="Radios${i}"]:checked`).value;

        let object = {
            icdDiagnosisId: await getIdFromICD10(icdDiagnosisId),
            description: description,
            type: type
        };

        diagnoses[`${i - 1}`] = object;
    }

    const inspection = {
        anamnesis: document.getElementById('anamnesis').value,
        complaints: document.getElementById('complaints').value,
        treatment: document.getElementById('treatment').value,
        conclusion: document.getElementById('conclusionSelect').value,
        nextVisitDate: document.getElementById('diseaseDate').value !== "" ? document.getElementById('diseaseDate').value : null,
        deathDate: document.getElementById('deathDate').value !== "" ? document.getElementById('deathDate').value : null,
        diagnoses: diagnoses,
    }
    console.log(inspection);

    putInspection(inspection, token);
}

async function putInspection(data, token) {
    const url = `https://mis-api.kreosoft.space/api/inspection/${inspectionId}`;
    console.log(inspectionId);
    return fetch(url, {
        method: 'PUT',
        body: JSON.stringify(data),
        headers: new Headers({
            'Content-Type': 'application/json',
            "Authorization": `Bearer ${token}`
        })
    })
        .then(response => response.json())
        .then(result => {
            console.log(result);
            const errorMessage = document.getElementById('errorMessage');
            errorMessage.textContent = '';

            if (result.message) {
                errorMessage.textContent = result.message;
                console.log(result.message);
            }

            if (result.title) {
                errorMessage.textContent = result.title;
                console.log(result.title);
            }

            if (result.token) {
                token = result.token;
                localStorage.setItem('token', token);
                window.location.href = '../patients/patients.html';
            }

            console.log(result);
        })
        .catch(error => {
            //console.error('Ошибка', error);
            //const errorMessage = document.getElementById('errorMessage');
            //errorMessage.textContent = 'Произошла ошибка при создании пациента. Пожалуйста, попробуйте еще раз.';
            console.error('Ошибка', error);
            window.location.reload(true);
        });
}