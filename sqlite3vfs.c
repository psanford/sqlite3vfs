#include "sqlite3vfs.h"
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#ifdef SQLITE3VFS_LOADABLE_EXT
SQLITE_EXTENSION_INIT1
#endif

extern int goVFSOpen(sqlite3_vfs* vfs, const char * name, sqlite3_file* file, int flags, int *outFlags);
extern int goVFSDelete(sqlite3_vfs*, const char *zName, int syncDir);
extern int goVFSAccess(sqlite3_vfs*, const char *zName, int flags, int* pResOut);
extern int goVFSFullPathname(sqlite3_vfs*, const char *zName, int nOut, char* zOut);
extern int goVFSRandomness(sqlite3_vfs*, int nByte, char *zOut);
extern int goVFSSleep(sqlite3_vfs*, int microseconds);
extern int goVFSCurrentTimeInt64(sqlite3_vfs*, sqlite3_int64* piNow);

extern int goVFSClose(sqlite3_file* file);
extern int goVFSRead(sqlite3_file* file, void* buf, int iAmt, sqlite3_int64 iOfst);
extern int goVFSWrite(sqlite3_file* file ,const void* buf, int iAmt, sqlite3_int64 iOfst);
extern int goVFSTruncate(sqlite3_file* file, sqlite3_int64 size);
extern int goVFSSync(sqlite3_file* file, int flags);
extern int goVFSFileSize(sqlite3_file* file, sqlite3_int64 *pSize);
extern int goVFSLock(sqlite3_file* file, int eLock);
extern int goVFSUnlock(sqlite3_file*, int eLock);
extern int goVFSCheckReservedLock(sqlite3_file* file, int *pResOut);
extern int goVFSSectorSize(sqlite3_file* file);
extern int goVFSDeviceCharacteristics(sqlite3_file* file);
extern int goVFSFileControl(sqlite3_file* file, int op, void *pArg);


int s3vfsNew(char* name, int maxPathName) {
  sqlite3_vfs* vfs;
  sqlite3_vfs* delegate;
  vfs = calloc(1, sizeof(sqlite3_vfs));
  if (vfs == NULL) {
    return SQLITE_ERROR;
  }

  delegate = sqlite3_vfs_find(0);

  vfs->iVersion = 2;
  vfs->szOsFile = sizeof(s3vfsFile);
  vfs->mxPathname = maxPathName;
  vfs->zName = name;
  vfs->xOpen = s3vfsOpen;
  vfs->xDelete = s3vfsDelete;
  vfs->xAccess = s3vfsAccess;
  vfs->xFullPathname = s3vfsFullPathname;
  vfs->xDlOpen = delegate->xDlOpen;
  vfs->xDlError = delegate->xDlError;
  vfs->xDlSym = delegate->xDlSym;
  vfs->xDlClose = delegate->xDlClose;
  vfs->xRandomness = s3vfsRandomness;
  vfs->xSleep = s3vfsSleep;
  vfs->xCurrentTime = s3vfsCurrentTime;
  vfs->xGetLastError = delegate->xGetLastError;
  vfs->xCurrentTimeInt64 = s3vfsCurrentTimeInt64;

  return sqlite3_vfs_register(vfs, 0);
}

int s3vfsOpen(sqlite3_vfs* vfs, const char * name, sqlite3_file* file, int flags, int *outFlags) {
  int ret = goVFSOpen(vfs, name, file, flags, outFlags);
  file->pMethods = &s3vfs_io_methods;
  return ret;
}

int s3vfsDelete(sqlite3_vfs* vfs, const char *zName, int syncDir) {
  int ret = goVFSDelete(vfs, zName, syncDir);
  return ret;
}

int s3vfsAccess(sqlite3_vfs* vfs, const char *zName, int flags, int* pResOut) {
  return goVFSAccess(vfs, zName, flags, pResOut);
}

int s3vfsFullPathname(sqlite3_vfs* vfs, const char *zName, int nOut, char* zOut) {
  return goVFSFullPathname(vfs, zName, nOut, zOut);
}

int s3vfsRandomness(sqlite3_vfs* vfs, int nByte, char *zOut) {
  return goVFSRandomness(vfs, nByte, zOut);
}

int s3vfsSleep(sqlite3_vfs* vfs, int microseconds) {
  return goVFSSleep(vfs, microseconds);
}

int s3vfsCurrentTime(sqlite3_vfs* vfs, double* prNow) {
  sqlite3_int64 i = 0;
  int rc;
  rc = s3vfsCurrentTimeInt64(0, &i);
  *prNow = i/86400000.0;
  return rc;
}

int s3vfsCurrentTimeInt64(sqlite3_vfs* vfs, sqlite3_int64* piNow) {
  return goVFSCurrentTimeInt64(vfs, piNow);
}

int s3vfsClose(sqlite3_file* file) {
  return goVFSClose(file);
}

int s3vfsRead(sqlite3_file* file, void* zBuf, int iAmt, sqlite3_int64 iOfst) {
  return goVFSRead(file, zBuf, iAmt, iOfst);
}

int s3vfsWrite(sqlite3_file* file, const void* zBuf, int iAmt, sqlite3_int64 iOfst) {
  return goVFSWrite(file, zBuf, iAmt, iOfst);
}

int s3vfsTruncate(sqlite3_file* file, sqlite3_int64 size) {
  return goVFSTruncate(file, size);
}

int s3vfsSync(sqlite3_file* file, int flags) {
  return goVFSSync(file, flags);
}

int s3vfsFileSize(sqlite3_file* file, sqlite3_int64 *pSize) {
  return goVFSFileSize(file, pSize);
}

int s3vfsLock(sqlite3_file* file, int eLock) {
  return goVFSLock(file, eLock);
}

int s3vfsUnlock(sqlite3_file* file , int eLock) {
  return goVFSUnlock(file, eLock);
}

int s3vfsCheckReservedLock(sqlite3_file* file, int *pResOut) {
  return goVFSCheckReservedLock(file, pResOut);
}

int s3vfsSectorSize(sqlite3_file* file) {
  return goVFSSectorSize(file);

}

int s3vfsDeviceCharacteristics(sqlite3_file* file) {
  return goVFSDeviceCharacteristics(file);
}

int s3vfsFileControl(sqlite3_file *pFile, int op, void *pArg){
  return goVFSFileControl(pFile, op, pArg);
}

char* s3vfsSqlite3Mprintf(const char* str) {
  return sqlite3_mprintf("%s", str);
}

const char* s3vfsURIKey(const char* zName, int n) {
  return zName == NULL ? NULL : sqlite3_uri_key(zName, n);
}

const char* s3vfsURIParameter(const char* zName, const char* zParam) {
  return zName == NULL ? NULL : sqlite3_uri_parameter(zName, zParam);
}

const sqlite3_io_methods s3vfs_io_methods = {
  1,                               /* iVersion */
  s3vfsClose,                      /* xClose */
  s3vfsRead,                       /* xRead */
  s3vfsWrite,                      /* xWrite */
  s3vfsTruncate,                   /* xTruncate */
  s3vfsSync,                       /* xSync */
  s3vfsFileSize,                   /* xFileSize */
  s3vfsLock,                       /* xLock */
  s3vfsUnlock,                     /* xUnlock */
  s3vfsCheckReservedLock,          /* xCheckReservedLock */
  s3vfsFileControl,                /* xFileControl */
  s3vfsSectorSize,                 /* xSectorSize */
  s3vfsDeviceCharacteristics,      /* xDeviceCharacteristics */
};

/* Wrappers making a registered sqlite3_vfs's function pointers callable from
** Go, so a Go VFS can delegate the files it cannot serve. */

sqlite3_vfs* s3vfsFind(const char* zName) {
  return sqlite3_vfs_find(zName);
}

int s3vfsBaseOsFileSize(sqlite3_vfs* vfs) {
  return vfs->szOsFile;
}

int s3vfsBaseMaxPathname(sqlite3_vfs* vfs) {
  return vfs->mxPathname;
}

/* sqlite3_malloc, not Go memory: sqlite holds this pointer for the life of
** the file. */
sqlite3_file* s3vfsBaseFileAlloc(sqlite3_vfs* vfs) {
  sqlite3_file* file = (sqlite3_file*)sqlite3_malloc(vfs->szOsFile);
  if (file != NULL) {
    memset(file, 0, vfs->szOsFile);
  }
  return file;
}

void s3vfsBaseFileFree(sqlite3_file* file) {
  sqlite3_free(file);
}

int s3vfsBaseOpen(sqlite3_vfs* vfs, const char* zName, sqlite3_file* file, int flags, int* outFlags) {
  return vfs->xOpen(vfs, zName, file, flags, outFlags);
}

int s3vfsBaseDelete(sqlite3_vfs* vfs, const char* zName, int syncDir) {
  return vfs->xDelete(vfs, zName, syncDir);
}

int s3vfsBaseAccess(sqlite3_vfs* vfs, const char* zName, int flags, int* pResOut) {
  return vfs->xAccess(vfs, zName, flags, pResOut);
}

int s3vfsBaseFullPathname(sqlite3_vfs* vfs, const char* zName, int nOut, char* zOut) {
  return vfs->xFullPathname(vfs, zName, nOut, zOut);
}

int s3vfsBaseClose(sqlite3_file* file) {
  if (file->pMethods == NULL) {
    return SQLITE_OK;
  }
  return file->pMethods->xClose(file);
}

int s3vfsBaseRead(sqlite3_file* file, void* buf, int iAmt, sqlite3_int64 iOfst) {
  return file->pMethods->xRead(file, buf, iAmt, iOfst);
}

int s3vfsBaseWrite(sqlite3_file* file, const void* buf, int iAmt, sqlite3_int64 iOfst) {
  return file->pMethods->xWrite(file, buf, iAmt, iOfst);
}

int s3vfsBaseTruncate(sqlite3_file* file, sqlite3_int64 size) {
  return file->pMethods->xTruncate(file, size);
}

int s3vfsBaseSync(sqlite3_file* file, int flags) {
  return file->pMethods->xSync(file, flags);
}

int s3vfsBaseFileSize(sqlite3_file* file, sqlite3_int64* pSize) {
  return file->pMethods->xFileSize(file, pSize);
}

int s3vfsBaseLock(sqlite3_file* file, int eLock) {
  return file->pMethods->xLock(file, eLock);
}

int s3vfsBaseUnlock(sqlite3_file* file, int eLock) {
  return file->pMethods->xUnlock(file, eLock);
}

int s3vfsBaseCheckReservedLock(sqlite3_file* file, int* pResOut) {
  return file->pMethods->xCheckReservedLock(file, pResOut);
}

int s3vfsBaseSectorSize(sqlite3_file* file) {
  return file->pMethods->xSectorSize(file);
}

int s3vfsBaseDeviceCharacteristics(sqlite3_file* file) {
  return file->pMethods->xDeviceCharacteristics(file);
}
