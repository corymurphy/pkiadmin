document.addEventListener("DOMContentLoaded", (event) => {
    document.body.addEventListener('htmx:beforeSwap', function(evt) {
        if (evt.detail.xhr.status === 422) {
                console.log("setting status to paint");
            // allow 422 responses to swap as we are using this as a signal that
            // a form was submitted with bad data and want to rerender with the
            // errors
            //
            // set isError to false to avoid error logging in console
            evt.detail.shouldSwap = true;
            evt.detail.isError = false;
        }
    });
            });