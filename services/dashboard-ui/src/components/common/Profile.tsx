'use client'

import { GoMail } from "react-icons/go"
import { useUser } from '@auth0/nextjs-auth0/client'
import { Text } from '@/components'

export const Profile = () => {
  const { user, error, isLoading } = useUser()

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>{error.message}</div>

  return (
    user && (
      <div className="flex gap-4 items-center">
        <img
          className="rounded-full"
          height="40px"
          width="40px"
          src={user.picture as string}
          alt={user.name as string}
        />
        <div className="flex flex-col gap-0">
          <Text className="truncate" variant="label">{user.name}</Text>
          <Text className="truncate" variant="caption">{user.email}</Text>
        </div>
      </div>
    )
  )
}


export const ProfileDropdown = () => {
  const { user, error, isLoading } = useUser()
  if (isLoading) return <div>Loading...</div>
  if (error) return <div>{error.message}</div>
  
  return user && (
    <>
      <style>
        {`.dropdown:focus-within .dropdown-menu {
  display:block;
}`}
      </style>
      <div className="z-10 relative inline-block text-left dropdown">
        <span className="rounded shadow-sm">
          <button
            className="inline-flex justify-center w-full px-4 py-2 text-sm transition duration-150 ease-in-out bg-slate-50 dark:bg-slate-950 rounded"
            type="button"
            aria-haspopup="true"
            aria-expanded="true"
            aria-controls="headlessui-menu-items-117"
          >
            <Profile />
            <svg
              className="w-5 h-5 ml-2 -mr-1"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                clipRule="evenodd"
              ></path>
            </svg>
          </button>
        </span>
        <div className="hidden dropdown-menu">
          <div
            className="absolute right-0 w-56 mt-2 origin-top-right bg-slate-50 dark:bg-slate-900 border divide-y rounded shadow-md outline-none"
            aria-labelledby="headlessui-menu-button-1"
            id="headlessui-menu-items-117"
            role="menu"
          >
            <div className="px-4 py-3">
              <p className="text-sm leading-5">Signed in as</p>              
              <p className="text-sm font-medium leading-5 truncate">
                {user && user?.email}
              </p>
            </div>
            <div className="">
              {/* <a
                  href="javascript:void(0)"
                  className="text-slate-700 flex justify-between w-full px-4 py-2 text-sm leading-5 text-left"
                  role="menuitem"
                  >
                  Account settings
                  </a> */}
              <a
                href="mailto:team@nuon.co"
                className="hover:bg-slate-100 dark:hover:bg-slate-800 flex gap-2 items-center w-full px-4 py-2 text-sm leading-5 text-left"
                role="menuitem"
              >
                <GoMail /> Support
              </a>
            </div>
            <div className="">
              <a
                href="/api/auth/logout"
                className="hover:bg-slate-100 dark:hover:bg-slate-800 flex justify-between w-full px-4 py-2 text-sm leading-5 text-left"
                role="menuitem"
              >
                Sign out
              </a>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
