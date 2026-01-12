import { updatePagination, setPage } from "../patient.js";


async function GeneralTestFunction(Page, maxPagination){
    setPage(Page)
    updatePagination(maxPagination)
}


document.getElementById('submitButton').addEventListener('click', function() {
    const value1 = parseFloat(document.getElementById('input1').value);
    const value2 = parseFloat(document.getElementById('input2').value);
    GeneralTestFunction(value1, value2);
});


