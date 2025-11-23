const url = new URL(window.location.href);
console.log(url);
const params = new URLSearchParams(url.search);
var token = localStorage.getItem('token');


const previousInspectionId = params.get('previousInspectionId');
const patient = params.get('patient');
const repeat = params.get('repeat');
if (repeat === 'true') {
    document.getElementById('checkedRadioSwitch').checked = true;
} else {
    document.getElementById('checkedRadioSwitch').checked = false;
    console.log(repeat);
}

console.log(previousInspectionId)

const checkedRadioSwitch = document.getElementById('checkedRadioSwitch');
const selectContainer = document.getElementById('selectContainer');

if (repeat === 'true' || checkedRadioSwitch.checked) {
    selectContainer.style.display = 'block';
} else {
    selectContainer.style.display = 'none';
}

checkedRadioSwitch.addEventListener('change', function () {
    if (this.checked) {
        selectContainer.style.display = 'block';
    } else {
        selectContainer.style.display = 'none';
        selectPreviousInspections.value = '';
    }
});

let diagnosCount = 1;
let commentCount = 1;
let numberOfObjects1 = 1;
let numberOfObjects2 = 1;
let id = null;
const checkConsultation = document.getElementById('checkConsultation');
const selectSpecialties = document.getElementById('selectSpecialties1');
const consultationComment = document.getElementById('1ConsultationComment');
const addConsultation = document.getElementById('addConsultation');

checkConsultation.addEventListener('change', function () {
    if (this.checked) {
        selectSpecialties.style.display = 'block';
        consultationComment.style.display = 'block';
        addConsultation.style.display = 'block';
    } else {
        selectSpecialties.style.display = 'none';
        consultationComment.style.display = 'none';
        addConsultation.style.display = 'none';

        const consultations = document.querySelectorAll('[id^="consultation"]');
        for (let i = 2; i < consultations.length; i++) {
            consultations[i].remove();
        }
        numberOfObjects1 = 1;
    }
});


const urlSpecialities = `https://mis-api.kreosoft.space/api/dictionary/speciality?size=30`;
getSpecialities(urlSpecialities, numberOfObjects1);

const urlPreviousInspections = `https://mis-api.kreosoft.space/api/patient/${patient}/inspections/search`;
getPreviousInspections(urlPreviousInspections, token)
    .then(() => {
        getPreviousInspection1();
    })
    .catch(error => {
        console.error('Ошибка', error);
    });

const urlPreviousInspection = `https://mis-api.kreosoft.space/api/inspection/${previousInspectionId}`;
getPreviousInspection(urlPreviousInspection, token);

function getPreviousInspection1() {
    const optionToSelect = document.getElementById('selectPreviousInspections');//.querySelector(`option[value="${previousInspectionId}"]`);
    console.log(optionToSelect);
    console.log(previousInspectionId);
    console.log(optionToSelect.querySelector(`option[value="${previousInspectionId}"]`));
    const selectedOption = optionToSelect.querySelector(`option[value="${previousInspectionId}"]`);
    if (optionToSelect) {
        optionToSelect.selectedIndex = selectedOption.index;
        console.log(optionToSelect.selected);
    }
}


const urlPatient = `https://mis-api.kreosoft.space/api/patient/${patient}`;
getPatient(urlPatient, token);


async function getPatient(url, token) {
    return fetch(url, {
        method: 'GET',
        headers: new Headers({
            "Authorization": `Bearer ${token}`
        }),
    })
        .then(response => response.json())
        .then(async data => {
            console.log(data)
            document.querySelector('[data-name]').textContent = data.name;
            document.querySelector('[data-birthday]').textContent = `Дата рождения: ${data.birthday ? await formatBirthday(data.birthday.split('T')[0]) : 'Не указано'}`;
            document.querySelector('[data-gender]').textContent = data.gender === 'Male' ? '♂' : data.gender === 'Female' ? '♀' : '';
        })
        .catch(error => {
            console.error('Ошибка', error);
        });
}


async function getPreviousInspection(url, token) {
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
            const previousInspectionDate = formatDataTime(data.date);
            const selectPreviousInspections = document.getElementById('selectPreviousInspections');
            selectPreviousInspections.name = data.diagnoses.id;
            id = data.id;
        })
        .catch(error => {
            console.error('Ошибка', error);
        });
}


function getPreviousInspections(url, token) {
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
            populateInspections(data);
        })
        .catch(error => {
            console.error('Ошибка', error);
        });
}


async function getSpecialities(url, i) {
    return fetch(url, {
        method: 'GET',
    })
        .then(response => response.json())
        .then(data => {
            console.log(url);
            console.log(data);
            populateSpecialties(data.specialties, i);
        })
        .catch(error => {
            console.error('Ошибка', error);
        });
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


async function post(url, data, token) {
    return fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify(data)
    })
        .then(response => response.json())
        .then(result => {
            const errorMessage = document.getElementById('errorMessage');
            errorMessage.textContent = '';
            console.log(result);

            if (result.message) {
                errorMessage.textContent = result.message;
                console.log(result.message);
            }
            if (result.title) {
                errorMessage.textContent = result.title;
                console.log(result.title);
            }
        })
        .catch(error => {
            console.error('Ошибка', error);
            const errorMessage = document.getElementById('errorMessage');
            errorMessage.textContent = 'Произошла ошибка. Пожалуйста, попробуйте еще раз.';
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



function populateInspections(inspections) {
    const selectInspections = document.getElementById('selectPreviousInspections');
    inspections.forEach(inspection => {
        const option = document.createElement('option');
        option.value = inspection.id;
        option.text = inspection.diagnosis.code + ' ' + formatDataTime(inspection.date);
        selectInspections.appendChild(option);
        console.log(inspections);
    });
}

function populateSpecialties(specialties, i) {
    const selectSpecialties = document.getElementById(`selectSpecialties${i}`);
    console.log(selectSpecialties);
    specialties.forEach(specialty => {
        const option = document.createElement('option');
        option.value = specialty.id;
        option.text = specialty.name;
        selectSpecialties.appendChild(option);
    });
}

function formatBirthday(originalDate) {
    const dateParts = originalDate.split("T")[0].split("-");
    const day = dateParts[2];
    const month = dateParts[1];
    const year = dateParts[0];
    const formattedDate = `${day}.${month}.${year}`;
    return formattedDate;
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


//------------------------Consultation--------------------

document.getElementById('addConsultationBtn').addEventListener('click', function () {
    numberOfObjects1++;
    const consultationId = numberOfObjects1;
    const newConsultation = document.createElement('div');
    newConsultation.id = consultationId;
    newConsultation.innerHTML = `
      <div class="d-flex align-items-center my-3">
        <div class="form-check form-switch col">
        </div>
        <div class="col">
          <select class="form-select" id="selectSpecialties${numberOfObjects1}" required>
          </select>
        </div>
      </div>
      <div id="${consultationId}ConsultationComment">
        <label class="form-label fw-light py-1">Комментарий</label>
        <textarea class="form-control" rows="3"></textarea>
      </div>
      <button class="btn btn-danger delete-consultation-btn my-1" data-consultation-id="${consultationId}">- Удалить консультацию</button>
    `;
    getSpecialities(urlSpecialities, numberOfObjects1);
    const consultationContainer = document.getElementById('consultation1');
    consultationContainer.parentNode.insertBefore(newConsultation, consultationContainer.nextSibling);

    numberOfObjects1 = document.querySelectorAll('.delete-consultation-btn').length+1;
});

document.addEventListener('click', function (event) {
    if (event.target.classList.contains('delete-consultation-btn')) {
        const consultationId = event.target.dataset.consultationId;
        const consultationElement = document.getElementById(consultationId);
        if (consultationElement) {
            consultationElement.remove();
        }
        numberOfObjects1 = document.querySelectorAll('.delete-consultation-btn').length+1;
    } 
   else if (event.target.classList.contains('delete-diagnos-btn')) {
        const diagnosId = event.target.dataset.diagnosId;
        const diagnosElement = document.getElementById(diagnosId);
        if (diagnosElement) {
            diagnosElement.remove();
        }
        numberOfObjects2 = document.querySelectorAll('.delete-diagnos-btn').length+1;
        console.log(numberOfObjects2);
    }
});

//------------------------Diagnosies--------------------
const inputElement1 = document.getElementById('selectDiagnoses1');
inputElement1.addEventListener('input', function () {
    const input = event.target.value;
    const url = `https://mis-api.kreosoft.space/api/dictionary/icd10?request=${input}&page=1&size=5`;
    getDiagnosis(url, 1);
});


document.getElementById('addDiagnosBtn').addEventListener('click', function () {
    numberOfObjects2++;
    const diagnosId = numberOfObjects2;
    const newDiagnos = document.createElement('div');
    newDiagnos.id = diagnosId;
    newDiagnos.innerHTML = `
        <div class="d-flex align-items-center">
            <div class="col">
                <label class="form-label fw-light py-1">Болезни</label>
                <input type="text" class="form-control" placeholder="" id="selectDiagnoses${numberOfObjects2}" list="diagnoses${numberOfObjects2}">
                <datalist id="diagnoses${numberOfObjects2}">
                </datalist>
            </div>
        </div>
        <div id="${diagnosId}DiagnosComment">
            <textarea class="form-control my-2" rows="1"></textarea>
        </div>
        <label class="form-label fw-light py-1">Тип диагноза в осмотре</label>
        <div class="form-group">
          <div class="form-check form-check-inline">
            <input class="form-check-input" type="radio" name="Radios${numberOfObjects2}" id="Main${numberOfObjects2}" value="Main">
            <label class="form-check-label">
              Основной
            </label>
          </div>
          <div class="form-check form-check-inline">
            <input class="form-check-input" type="radio" name="Radios${numberOfObjects2}" id="Concomitant${numberOfObjects2}" value="Concomitant" checked>
            <label class="form-check-label">
              Сопутствующий
            </label>
          </div>
          <div class="form-check form-check-inline">
            <input class="form-check-input" type="radio" name="Radios${numberOfObjects2}" id="Complication${numberOfObjects2}" value="Complication">
            <label class="form-check-label">
              Осложнение
            </label>
          </div>
        </div>
        <p id="diagnos1"></p>
        <div class="col my-2" id="addDiagnos">
            <button class="btn btn-danger delete-diagnos-btn my-1" data-diagnos-id="${numberOfObjects2}">- Удалить диагноз</button>
        </div>
    `;
    const diagnosContainer = document.getElementById('diagnos1');
    diagnosContainer.parentNode.insertBefore(newDiagnos, diagnosContainer.nextSibling);

    const inputElement = document.getElementById(`selectDiagnoses${numberOfObjects2}`);
    inputElement.addEventListener('input', function () {
        const input = event.target.value;
        const url = `https://mis-api.kreosoft.space/api/dictionary/icd10?request=${input}&page=1&size=5`;
        getDiagnosis(url, numberOfObjects2);
    });
    ;
    numberOfObjects2 = document.querySelectorAll('.delete-diagnos-btn').length+1;
});

// -----------------------Заключения
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

const saveInspectionBtn = document.getElementById('saveInspectionBtn');
const cancelBtn = document.getElementById('cancelBtn');


saveInspectionBtn.addEventListener('click', function () {
    InspectionCreate();
    console.log('Осмотр сохранен');
});

cancelBtn.addEventListener('click', function () {
    console.log('Создание осмотра отменено');
    window.location.href = `../patient/patient.html?${patient}=`;
});



async function InspectionCreate() {
    let diagnoses = [];
    console.log(numberOfObjects2);
    for (let i = 1; i <= numberOfObjects2; i++) {
        console.log(i);
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

    let consultations = [];
    for (let i = 1; i <= numberOfObjects1; i++) {
        let specialityId = document.getElementById(`selectSpecialties${i}`).value;
        let commentContent = document.getElementById(`${i}ConsultationComment`).querySelector('textarea').value;

        let comment = {
            content: commentContent
        };

        let consultation = {
            specialityId: specialityId,
            comment: comment
        };

        consultations.push(consultation);
    }

    const inspection = {
        date: document.getElementById('inputData').value,
        anamnesis: document.getElementById('anamnesis').value,
        complaints: document.getElementById('complaints').value,
        treatment: document.getElementById('treatment').value,
        conclusion: document.getElementById('conclusionSelect').value,
        nextVisitDate: document.getElementById('diseaseDate').value !== "" ? document.getElementById('diseaseDate').value : null,
        deathDate: document.getElementById('deathDate').value !== "" ? document.getElementById('deathDate').value : null,
        previousInspectionId: id,
        diagnoses: diagnoses,
        consultations: consultations,
    }
    console.log(inspection);

    const urlPost = `https://mis-api.kreosoft.space/api/patient/${patient}/inspections`
    post(urlPost, inspection, token);
}

const selectElement = document.getElementById('selectPreviousInspections');
selectElement.addEventListener('change', () => {
    const selectedOption = selectElement.options[selectElement.selectedIndex];
    id = selectedOption.value;
    console.log(id);
});