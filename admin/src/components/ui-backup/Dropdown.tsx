import { Fragment } from 'react';
import type { ReactNode } from 'react';
import { Menu, Transition } from '@headlessui/react';

interface DropdownProps {
  trigger: ReactNode;
  children: ReactNode;
  align?: 'left' | 'right';
}

export function Dropdown({ trigger, children, align = 'right' }: DropdownProps) {
  const alignStyles = {
    left: 'left-0',
    right: 'right-0',
  };

  return (
    <Menu as="div" className="relative inline-block text-left">
      <Menu.Button as={Fragment}>{trigger}</Menu.Button>

      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items
          className={`absolute ${alignStyles[align]} mt-2 w-56 origin-top-right rounded-lg bg-white dark:bg-gray-800 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none z-10`}
        >
          <div className="py-1">{children}</div>
        </Menu.Items>
      </Transition>
    </Menu>
  );
}

interface DropdownItemProps {
  children: ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  icon?: ReactNode;
}

export function DropdownItem({ children, onClick, disabled, icon }: DropdownItemProps) {
  return (
    <Menu.Item disabled={disabled}>
      {({ active }) => (
        <button
          onClick={onClick}
          className={`
            ${active ? 'bg-gray-100 dark:bg-gray-700' : ''}
            ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
            group flex w-full items-center px-4 py-2 text-sm text-gray-700 dark:text-gray-200
          `}
        >
          {icon && <span className="mr-3">{icon}</span>}
          {children}
        </button>
      )}
    </Menu.Item>
  );
}

export function DropdownDivider() {
  return <div className="my-1 h-px bg-gray-200 dark:bg-gray-700" />;
}
