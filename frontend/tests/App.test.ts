import { mount } from '@vue/test-utils';
import { describe, it, expect } from 'vitest';
import App from '../src/App.vue';

describe('Landing Page', () => {
    it('renders header', () => {
        const wrapper = mount(App);
        expect(wrapper.text()).toContain('Arch Stats');
    });
});
