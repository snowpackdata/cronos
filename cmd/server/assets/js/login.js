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
        const tenantSlug = result.tenant_slug;
        const isStaff = result.is_staff;
        
        // Determine the target URL based on current environment
        const currentHost = window.location.hostname;
        const currentProtocol = window.location.protocol;
        const currentPort = window.location.port;
        
        let targetHost;
        if (currentHost === 'localhost' || currentHost === '127.0.0.1') {
            // Local development - use tenant.localhost
            targetHost = `${tenantSlug}.localhost`;
        } else {
            // Production - use tenant.domain.com
            const baseDomain = currentHost.replace(/^[^.]+\./, ''); // Remove existing subdomain if any
            targetHost = `${tenantSlug}.${baseDomain}`;
        }
        
        const portPart = currentPort ? `:${currentPort}` : '';
        const redirectPath = isStaff ? '/admin/timesheet' : '/portal/dashboard';
        const redirectURL = `${currentProtocol}//${targetHost}${portPart}${redirectPath}?token=${token}`;
        
        window.location.href = redirectURL;
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


