document.addEventListener("DOMContentLoaded", function() {
      const checkbox = document.getElementById("agreeCheck");
      const submitBtn = document.getElementById("submitBtn");
      const nextBtn = document.getElementById("nextBtn");
      const backBtn = document.getElementById("backBtn");
      const passwordInpt = document.getElementById("password");
      const confirmPasswordInpt = document.getElementById("confirm-password");
      const messageBox = document.getElementById("message-box");

      // --- Validation Logic ---
      function validatePasswordStrength() {
          const value = passwordInpt.value;
          const requirements = [
              // Regex: At least 8 characters
              { id: 'req-length', regex: /.{8,}/ },
              // Regex: At least one lowercase letter
              { id: 'req-lower', regex: /(?=.*[a-z])/ },
              // Regex: At least one uppercase letter
              { id: 'req-upper', regex: /(?=.*[A-Z])/ },
              // Regex: At least one digit
              { id: 'req-number', regex: /(?=.*[0-9])/ },
              // Regex: At least one special character (common set)
              { id: 'req-symbol', regex: /(?=.*[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?])/ }
          ];

          let allValid = true;

          requirements.forEach(req => {
              const isValid = req.regex.test(value);
              const element = document.getElementById(req.id);

              if (element) {
                  // Toggle color class
                  element.classList.toggle('text-red-500', !isValid);
                  element.classList.toggle('text-blue-500', isValid);

                  // Update icon (using a simple path toggle for Check/X)
                  const iconPath = element.querySelector('svg path');
                  if (iconPath) {
                      if (isValid) {
                          // Checkmark path
                          iconPath.setAttribute('d', 'M4.5 12.75l6 6 9-13.5');
                          iconPath.setAttribute('stroke-width', '3'); // Make checkmark bolder
                      } else {
                          // X mark path
                          iconPath.setAttribute('d', 'M6 18L18 6M6 6l12 12');
                          iconPath.setAttribute('stroke-width', '2'); // Keep X standard width
                      }
                  }
              }

              if (!isValid) {
                  allValid = false;
              }
          });

          return allValid;
      }

      function validateConfirmPassword(isStrong) {
          const isMatch = passwordInpt.value === confirmPasswordInpt.value;

          // Remove all existing border classes

          if ((isMatch && isStrong) && confirmPasswordInpt.value.length > 0) {
              // Passwords match AND are strong -> GREEN border
              confirmPasswordInpt.classList.add("outline-blue-600");
              confirmPasswordInpt.classList.add("text-blue-600");
              confirmPasswordInpt.classList.add("border-blue-600");
              confirmPasswordInpt.classList.remove("outline-red-600");
              confirmPasswordInpt.classList.remove("text-red-600");
              confirmPasswordInpt.classList.remove("border-gray-400");

          } else if (confirmPasswordInpt.value.length > 0) {
              // Passwords don't match or aren't strong (but input is not empty) -> RED border
              confirmPasswordInpt.classList.add("outline-red-600");
              confirmPasswordInpt.classList.add("text-red-600");
              confirmPasswordInpt.classList.remove("outline-blue-600");
              confirmPasswordInpt.classList.remove("text-blue-600");
              confirmPasswordInpt.classList.remove("border-blue-600");

          } else {
              // Input is empty -> Default gray border
              confirmPasswordInpt.classList.add("border-gray-400");
          }

          return isMatch;
      }

      // --- Combined Handler ---
      function handleValidation() {
          const isStrong = validatePasswordStrength();
          const isMatch = validateConfirmPassword(isStrong);

          passwordInpt.classList.toggle("text-blue-600", isStrong)
          passwordInpt.classList.toggle("border-blue-600", isStrong)
          passwordInpt.classList.toggle("border-gray-400", !isStrong)

          // Enable or disable the registration button
          submitBtn.disabled = !(isStrong && isMatch);
      }

      passwordInpt.addEventListener("input", handleValidation)
      confirmPasswordInpt.addEventListener("input", handleValidation)

      nextBtn.addEventListener("click", function() {
          nextBtn.classList.add("hidden")
          document.getElementById("first-inpts").classList.add("hidden")
          document.getElementById("second-inpts").classList.remove("hidden")
          document.getElementById("second-btns").classList.remove("hidden")
      })

      backBtn.addEventListener("click", function() {
          nextBtn.classList.remove("hidden")
          document.getElementById("first-inpts").classList.remove("hidden")
          document.getElementById("second-inpts").classList.add("hidden")
          document.getElementById("second-btns").classList.add("hidden")
      })
})

