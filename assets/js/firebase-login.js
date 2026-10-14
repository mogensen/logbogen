import { initializeApp } from 'https://www.gstatic.com/firebasejs/10.14.1/firebase-app.js';
import {
    getAuth,
    connectAuthEmulator,
    GoogleAuthProvider,
    signInWithPopup,
    signInWithEmailAndPassword,
    createUserWithEmailAndPassword,
    sendEmailVerification,
    sendPasswordResetEmail,
    updateProfile,
    signOut,
} from 'https://www.gstatic.com/firebasejs/10.14.1/firebase-auth.js';

const root = document.getElementById('firebase-login');
const auth = getAuth(initializeApp({
    apiKey: root.dataset.apiKey,
    authDomain: root.dataset.authDomain,
    projectId: root.dataset.projectId,
}));

if (root.dataset.emulatorHost) {
    connectAuthEmulator(auth, 'http://' + root.dataset.emulatorHost, { disableWarnings: true });
}

const message = document.getElementById('login-message');
const form = document.getElementById('email-form');
const nameGroup = document.getElementById('name-group');
const nameInput = document.getElementById('login-name');
const emailInput = document.getElementById('login-email');
const passwordInput = document.getElementById('login-password');
const submitButton = document.getElementById('email-submit');
const toggleSignup = document.getElementById('toggle-signup');
const togglePrompt = document.getElementById('toggle-prompt');
const resendGroup = document.getElementById('resend-group');

let signingUp = false;
// The user whose verification mail can be sent again, kept after signing out
let unverified = null;

const firebaseErrors = {
    'auth/invalid-credential': 'Forkert e-mail eller adgangskode.',
    'auth/wrong-password': 'Forkert e-mail eller adgangskode.',
    'auth/user-not-found': 'Forkert e-mail eller adgangskode.',
    'auth/invalid-email': 'Skriv en gyldig e-mailadresse.',
    'auth/email-already-in-use': 'Der findes allerede en konto med den e-mail.',
    'auth/weak-password': 'Adgangskoden skal være mindst 6 tegn.',
    'auth/too-many-requests': 'For mange forsøg. Prøv igen senere.',
    'auth/network-request-failed': 'Netværksfejl. Tjek din forbindelse og prøv igen.',
};

const sessionErrors = {
    email_not_verified: 'Din e-mail er ikke bekræftet endnu. Tjek din indbakke.',
    stale_login: 'Dit login er for gammelt. Log ind igen.',
    invalid_token: 'Login mislykkedes. Prøv igen.',
    invalid_request: 'Login mislykkedes. Prøv igen.',
};

function show(text, kind) {
    message.textContent = text;
    message.className = 'alert mt-3 alert-' + kind;
}

function clearMessage() {
    message.className = 'alert d-none mt-3';
    message.textContent = '';
}

function busy(isBusy) {
    document.querySelectorAll('#firebase-login button').forEach(function (button) {
        button.disabled = isBusy;
    });
}

// Sends the ID token to the server, which starts the session when it verifies
async function startSession(user) {
    const response = await fetch('/auth/session', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Csrf-Token': window.csrfToken() },
        body: JSON.stringify({ idToken: await user.getIdToken() }),
    });
    const body = await response.json().catch(function () { return {}; });
    if (!response.ok) {
        const error = new Error(body.error || 'invalid_token');
        error.sessionError = body.error || 'invalid_token';
        throw error;
    }
    await signOut(auth);
    window.location.href = body.redirect || '/';
}

function fail(error) {
    if (error.sessionError) {
        show(sessionErrors[error.sessionError] || sessionErrors.invalid_token, 'danger');
    } else {
        show(firebaseErrors[error.code] || 'Noget gik galt. Prøv igen.', 'danger');
    }
}

// A signed-in user must verify the email before the server accepts the login
async function requireVerified(user) {
    if (user.emailVerified) {
        return startSession(user);
    }
    unverified = user;
    resendGroup.classList.remove('d-none');
    show('Din e-mail er ikke bekræftet endnu. Tjek din indbakke og log ind igen bagefter.', 'warning');
}

async function run(action) {
    clearMessage();
    resendGroup.classList.add('d-none');
    busy(true);
    try {
        await action();
    } catch (error) {
        fail(error);
    } finally {
        busy(false);
    }
}

document.getElementById('google-login').addEventListener('click', function () {
    run(async function () {
        try {
            const result = await signInWithPopup(auth, new GoogleAuthProvider());
            await startSession(result.user);
        } catch (error) {
            if (error.code === 'auth/popup-closed-by-user' || error.code === 'auth/cancelled-popup-request') {
                window.location.href = '/';
                return;
            }
            throw error;
        }
    });
});

form.addEventListener('submit', function (event) {
    event.preventDefault();
    run(async function () {
        const email = emailInput.value.trim();
        const password = passwordInput.value;
        if (!email || !password) {
            show('Skriv både e-mail og adgangskode.', 'danger');
            return;
        }

        if (signingUp) {
            const credential = await createUserWithEmailAndPassword(auth, email, password);
            const name = nameInput.value.trim();
            if (name) {
                await updateProfile(credential.user, { displayName: name });
            }
            await sendEmailVerification(credential.user);
            await signOut(auth);
            show('Vi har sendt en bekræftelsesmail til ' + email + '. Bekræft adressen og log ind bagefter.', 'success');
            setSigningUp(false);
            return;
        }

        const credential = await signInWithEmailAndPassword(auth, email, password);
        await requireVerified(credential.user);
    });
});

document.getElementById('resend-verification').addEventListener('click', function () {
    run(async function () {
        await sendEmailVerification(unverified);
        show('Vi har sendt bekræftelsesmailen igen.', 'success');
    });
});

document.getElementById('forgot-password').addEventListener('click', function (event) {
    event.preventDefault();
    run(async function () {
        const email = emailInput.value.trim();
        if (!email) {
            show('Skriv din e-mail først, så sender vi et link til at nulstille adgangskoden.', 'danger');
            return;
        }
        try {
            await sendPasswordResetEmail(auth, email);
        } catch (error) {
            // Whether an account exists is not revealed
            if (error.code !== 'auth/user-not-found') {
                throw error;
            }
        }
        show('Hvis der findes en konto med den e-mail, har vi sendt et link til at nulstille adgangskoden.', 'success');
    });
});

function setSigningUp(value) {
    signingUp = value;
    nameGroup.classList.toggle('d-none', !value);
    submitButton.textContent = value ? 'Opret konto' : 'Log ind';
    togglePrompt.textContent = value ? 'Har du allerede en logbog?' : 'Har du ikke en logbog endnu?';
    toggleSignup.textContent = value ? 'Log ind' : 'Opret en!';
    passwordInput.autocomplete = value ? 'new-password' : 'current-password';
}

toggleSignup.addEventListener('click', function (event) {
    event.preventDefault();
    clearMessage();
    resendGroup.classList.add('d-none');
    setSigningUp(!signingUp);
});

// A user left signed in by an earlier visit must not be logged in without asking
signOut(auth);
