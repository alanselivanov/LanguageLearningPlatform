document.addEventListener("DOMContentLoaded", function() {
    let faqQuestions = document.querySelectorAll("dt");
    faqQuestions.forEach(function(question) {
        let arrow = document.createElement("span");
        arrow.textContent = " ▼";
        arrow.style.float = "right";
        question.appendChild(arrow);

        question.style.cursor = "pointer";
        question.style.border = "1px solid #6ac585";
        question.style.padding = "10px";
        question.style.marginTop = "5px";
        question.style.transition = "all 0.3s";

        question.addEventListener("mouseover", function() {
            this.style.textDecoration = "underline";
        });
        question.addEventListener("mouseout", function() {
            this.style.textDecoration = "none";
        });

        question.addEventListener("click", function() {
            let answer = this.nextElementSibling;

            answer.classList.toggle("active");

            if (answer.classList.contains("active")) {
                answer.style.display = "block";
                arrow.textContent = " ▲";
            } else {
                answer.style.display = "none";
                arrow.textContent = " ▼";
            }
        });
    });

    let faqAnswers = document.querySelectorAll("dd");
    faqAnswers.forEach(function(answer) {
        answer.style.display = "none";
        answer.style.border = "1px solid #6ac585";
        answer.style.padding = "10px";
        answer.style.marginTop = "5px";
    });
});
