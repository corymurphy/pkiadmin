// document.addEventListener("DOMContentLoaded", (event) => {
//   document.body.addEventListener('htmx:beforeSwap', function (evt) {
//     if (evt.detail.xhr.status === 422) {
//       console.log("setting status to paint");
//       // allow 422 responses to swap as we are using this as a signal that
//       // a form was submitted with bad data and want to rerender with the
//       // errors
//       //
//       // set isError to false to avoid error logging in console
//       evt.detail.shouldSwap = true;
//       evt.detail.isError = false;
//     }
//   });
// });


// document.addEventListener('DOMContentLoaded', () => {
//   document.querySelectorAll('.fold-table tr.view').forEach(row => {
//     row.addEventListener('click', function() {
//       if (this.classList.contains('open')) {
//         this.classList.remove('open');
//         this.nextElementSibling.classList.remove('open');
//       } else {
//         document.querySelectorAll('.fold-table tr.view.open')
//           .forEach(openRow => {
//             openRow.classList.remove('open');
//             openRow.nextElementSibling.classList.remove('open');
//           });
//         this.classList.add('open');
//         this.nextElementSibling.classList.add('open');
//       }
//     });
//   });
// });

function toggleFold(e, rowId) {
  const event = e;
  if (event.preventDefault) event.preventDefault();
  thisRow = document.querySelector(`tr.view#${rowId}`)
  if (thisRow.classList.contains('open')) {
    thisRow.classList.remove('open');
    thisRow.nextElementSibling.classList.remove('open');
  } else {
    document.querySelectorAll('.fold-table tr.view.open')
      .forEach(openRow => {
        openRow.classList.remove('open');
        openRow.nextElementSibling.classList.remove('open');
      });
      thisRow.classList.add('open');
      thisRow.nextElementSibling.classList.add('open');
  }
}
