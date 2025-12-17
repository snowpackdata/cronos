const loginButton = document.getElementById('loginSubmit');
const loginEmail = document.getElementById('email');
const loginPassword = document.getElementById('password');
const loginError = document.getElementById('loginError');

loginButton.addEventListener('click', async (e) => {
    loginError.innerText = '';
    e.preventDefault();
    // Submit Login Form
    let postForm = new FormData();
    postForm.append('email', loginEmail.value);
    postForm.append('password', loginPassword.value);
    const requestOptions = {
      method: "POST",
      body: postForm,
      redirect: "follow"
    };
    fetch("/verify_login", requestOptions)
    .then((response) => response.text())
    .then((result) => {
        result = JSON.parse(result);
        if (result.status !== 200) {
            loginError.innerText = result.message;
            return;
        }
        let token = result.token;
        
        // Store the token in localStorage
        localStorage.setItem('snowpack_token', token);
        
        // Also store in a cookie for better server-side compatibility
        // Set cookie to expire in 30 days (same as the token)
        const expiryDate = new Date();
        expiryDate.setDate(expiryDate.getDate() + 30);
        
        // Get the current domain
        const domain = window.location.hostname;
        const isLocalhost = domain === 'localhost' || domain === '127.0.0.1';
        
        // Use domain specific settings only for non-localhost 
        const domainParam = isLocalhost ? '' : `; domain=${domain}`;
        
        // Set secure flag for HTTPS connections
        const secure = window.location.protocol === 'https:' ? '; Secure' : '';
        
        // Set the cookie with proper configuration for production
        document.cookie = `x-access-token=${token}; expires=${expiryDate.toUTCString()}; path=/${domainParam}${secure}; SameSite=Lax`;
        
        // Parse the token to check user role
        try {
            tokenJson = parseJwt(token);
            
            // Redirect based on user role
            if (tokenJson.IsStaff === true) {
                window.location.href = '/admin/timesheet';
            } else {
                window.location.href = '/portal/dashboard';
            }
        } catch (error) {
            // If we can't parse the token, default to 404 page
            window.location.href = '/404';
        }
    })
    .catch((error) => {
        loginError.innerText = "An error occurred during login. Please try again.";
    });
    return
});

function parseJwt (token) {
    var base64Url = token.split('.')[1];
    var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    var jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    return JSON.parse(jsonPayload);
}


