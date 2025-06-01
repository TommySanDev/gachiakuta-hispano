import { signal } from '@preact/signals';

// Message management for profile components
interface MessageState {
  success: {
    visible: boolean;
    message: string;
  };
  error: {
    visible: boolean;
    message: string;
  };
}

const messageStates = signal<Record<string, MessageState>>({});

// Helper to initialize message state for a component
function initializeMessageState(componentId: string) {
  if (!messageStates.value[componentId]) {
    messageStates.value = {
      ...messageStates.value,
      [componentId]: {
        success: { visible: false, message: '' },
        error: { visible: false, message: '' }
      }
    };
  }
}

// Profile messages manager
export const profileMessages = {
  showSuccess(componentId: string, message: string) {
    initializeMessageState(componentId);

    // Update DOM directly for immediate feedback
    const successDiv = document.getElementById(`${componentId}-success`);
    const successMessage = document.getElementById(`${componentId}-success-message`);
    const errorDiv = document.getElementById(`${componentId}-error`);

    if (successDiv && successMessage && errorDiv) {
      successMessage.textContent = message;
      successDiv.classList.remove('hidden');
      errorDiv.classList.add('hidden');
    }

    // Update signal state
    messageStates.value = {
      ...messageStates.value,
      [componentId]: {
        success: { visible: true, message },
        error: { visible: false, message: '' }
      }
    };
  },

  showError(componentId: string, message: string) {
    initializeMessageState(componentId);

    // Update DOM directly for immediate feedback
    const errorDiv = document.getElementById(`${componentId}-error`);
    const errorMessage = document.getElementById(`${componentId}-error-message`);
    const successDiv = document.getElementById(`${componentId}-success`);

    if (errorDiv && errorMessage && successDiv) {
      errorMessage.textContent = message;
      errorDiv.classList.remove('hidden');
      successDiv.classList.add('hidden');
    }

    // Update signal state
    messageStates.value = {
      ...messageStates.value,
      [componentId]: {
        success: { visible: false, message: '' },
        error: { visible: true, message }
      }
    };
  },

  hideMessages(componentId: string) {
    // Update DOM directly
    const successDiv = document.getElementById(`${componentId}-success`);
    const errorDiv = document.getElementById(`${componentId}-error`);

    if (successDiv && errorDiv) {
      successDiv.classList.add('hidden');
      errorDiv.classList.add('hidden');
    }

    // Update signal state
    if (messageStates.value[componentId]) {
      messageStates.value = {
        ...messageStates.value,
        [componentId]: {
          success: { visible: false, message: '' },
          error: { visible: false, message: '' }
        }
      };
    }
  }
};

// Loading states for different components
export const loadingStates = signal<Record<string, boolean>>({});

export const setLoading = (componentId: string, loading: boolean) => {
  loadingStates.value = {
    ...loadingStates.value,
    [componentId]: loading
  };
};

// Initialize the profile store
export function initializeProfileStore() {
  // Initialize common message states
  initializeMessageState('profile');
  initializeMessageState('password');
  initializeMessageState('security');

  console.log('✅ Profile store initialized');
}
