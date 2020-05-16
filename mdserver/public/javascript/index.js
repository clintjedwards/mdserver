function search(term) {
  if (term === "") {
    var table_rows = document.querySelectorAll("tr");
    table_rows.forEach(function (currentValue, currentIndex) {
      // skip the header table row
      if (currentIndex == 0) {
        return;
      }
      show(currentValue);
    });
    return;
  }

  fetch("/api/search?term=" + term)
    .then((response) => response.json())
    .then((hits) => {
      filterTable(hits);
    });
}

function filterTable(matches) {
  var table_rows = document.querySelectorAll("tr");
  table_rows.forEach(function (currentValue, currentIndex) {
    // skip the header table row
    if (currentIndex == 0) {
      return;
    }
    if (matches === null) {
      show(currentValue);
      return;
    }
    if (matches.includes(currentValue.id)) {
      show(currentValue);
      return;
    }
    console.log("id: " + currentValue.id);
    hide(currentValue);
    return;
  });
}

// Show an element
function show(elem) {
  elem.classList.remove("hidden");
}

// Hide an element
function hide(elem) {
  elem.classList.add("hidden");
}

document.addEventListener("DOMContentLoaded", function () {
  let search_bar = document.getElementById("search");
  search_bar.addEventListener(
    "keyup",
    debounce(function () {
      search(search_bar.value);
    }, 400)
  );
});
