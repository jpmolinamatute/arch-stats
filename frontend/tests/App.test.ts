import { mount } from '@vue/test-utils';
import { describe, it, expect } from 'vitest';
import App from '../src/App.vue';

describe('App.vue', () => {
    it('renders the default call to action when no view is active', () => {
        const wrapper = mount(App);
        expect(wrapper.text()).toContain('Open a session');
    });
});
