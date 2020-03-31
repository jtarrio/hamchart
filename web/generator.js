/*
   Copyright 2020 Jacobo Tarrío Barreiro (http://jacobo.tarrio.org)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Initialize the page.
function init() {
    enableDisableSubmit();
}

// Called when the user wants to enter coordinates directly.
function enterDirectly() {
    maidenheadEntry.style.display = "none";
    coordinateEntry.style.display = "block";
}

// Called when the user wants to enter a Maidenhead locator.
function enterMaidenhead() {
    maidenheadEntry.style.display = "block";
    coordinateEntry.style.display = "none";
}

// Decodes the Maidenhead locator and copies the coordinates to the "enter directly" box.
function copyMaidenhead() {
    let locator = locatorInput.value.trim().toLocaleLowerCase();
    let match = /^([a-r][a-r])([0-9][0-9])([a-x][a-x])?$/.exec(locator);
    let valid = match !== null;
    invalidLocatorLbl.style.display = valid ? "none" : "inline"
    if (!valid) {
        return;
    }

    let longitude = (match[1].codePointAt(0) - 97) * 20 - 180;
    let latitude = (match[1].codePointAt(1) - 97) * 10 - 90;
    longitude += (match[2].codePointAt(0) - 48) * 2;
    latitude += (match[2].codePointAt(1) - 48);
    let size = 1;
    if (match[3]) {
        longitude += (match[3].codePointAt(0) - 97) / 12;
        latitude += (match[3].codePointAt(1) - 97) / 24;
        size = 1 / 24;
    }

    longitudeInput.value = longitude + size;
    latitudeInput.value = latitude + size / 2;
    nameInput.value = "Maidenhead locator " + locator.substring(0, 2).toLocaleUpperCase() + locator.substring(2).toLocaleLowerCase();
    latitudeInputChanged();
    longitudeInputChanged();
    maidenheadEntry.style.display = "none";
    coordinateEntry.style.display = "block";
    nameInput.focus();
}

// Enables or disables the Submit button depending on whether the form is valid.
function enableDisableSubmit() {
    submitBtn.disabled = !validateInputs();
}

// Checks that the form input is valid.
function validateInputs() {
    return !!latitudeInput.value && !!longitudeInput.value && invalidLatitudeLbl.style.display === "none" && invalidLongitudeLbl.style.display === "none";
}

// Called when the user changes the contents of the latitude box.
function latitudeInputChanged() {
    let rewritten = rewriteCoordinate(latitudeInput.value, "N", "S", 90, -90)
    let invalid = rewritten === undefined;
    invalidLatitudeLbl.style.display = invalid ? "inline" : "none";
    if (!invalid && rewritten !== latitudeInput.value) {
        latitudeInput.value = rewritten;
    }
    enableDisableSubmit();
}

// Called when the user changes the contents of the longitude box.
function longitudeInputChanged() {
    let rewritten = rewriteCoordinate(longitudeInput.value, "E", "W", 180, -180)
    let invalid = rewritten === undefined;
    invalidLongitudeLbl.style.display = invalid ? "inline" : "none";
    if (!invalid && rewritten !== longitudeInput.value) {
        longitudeInput.value = rewritten;
    }
    enableDisableSubmit();
}

// When true, change the paper size when the 'metric' checkbox is toggled.
var tieMetricAndSize = true;

// Called when the 'metric' checkbox is toggled.
function metricBtnChanged() {
    if (tieMetricAndSize) {
        sizeSelect.value = metricBtn.checked ? "a4" : "letter";
        pageSizeSmartLbl.style.display = "block";
    }
}

// Called when the size select box is changed.
function sizeSelectChanged() {
    tieMetricAndSize = false;
}

// Rewrites a coordinate from deg-min-sec to a 4-digit decimal number.
function rewriteCoordinate(value, pos, neg, max, min) {
    let sgn = 0;
    let p = value.indexOf(pos);
    if (p >= 0) {
        sgn = 1;
        value = value.substring(0, p) + value.substring(p + 1);
    } else if ((p = value.indexOf(neg)) >= 0) {
        sgn = -1;
        value = value.substring(0, p) + value.substring(p + 1);
    }
    let numbers = value.split(/[^0-9.-]/).filter(n => !!n).map(Number.parseFloat);
    // We must have between 1 and 3 numbers.
    if (numbers.length < 1 || numbers.length > 3) {
        return;
    }
    // All numbers but the last must be integers.
    if (!numbers.every((number, index) => index == numbers.length - 1 || number == Math.floor(number))) {
        return;
    }
    // If the user specified a direction, the first number cannot be negative.
    if (sgn != 0 && numbers[0] < 0) {
        return;
    }
    // If the user didn't specify a direction, there must be only one number.
    if (sgn == 0 && numbers.length > 1) {
        return;
    }
    // The second and third number must lie between 0 and 60.
    if (numbers[1] < 0 || numbers[1] >= 60 || numbers[2] < 0 || numbers[2] >= 60) {
        return;
    }
    if (sgn == 0) {
        sgn = 1;
    }
    let coord = sgn * (numbers[0] + (numbers[1] || 0) / 60 + (numbers[2] || 0) / 3600);
    if (coord > max || coord < min) {
        return;
    }
    return Math.round(coord * 100000) / 100000;
}
